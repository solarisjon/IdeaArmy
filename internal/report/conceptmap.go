// Package report provides post-processing functions that enhance the generated
// HTML idea sheet, such as the interactive concept map.
package report

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/yourusername/ai-agent-team/internal/models"
)

// NodeType classifies a concept map node for styling purposes.
type NodeType string

const (
	NodeCenter    NodeType = "center"    // The winning idea
	NodePro       NodeType = "pro"       // A strength / supporting point
	NodeCon       NodeType = "con"       // A risk / challenge
	NodeIdea      NodeType = "idea"      // A runner-up idea
	NodeResearch  NodeType = "research"  // A research finding
	NodeImplement NodeType = "implement" // An implementation note
)

// ConceptMapNode is a single node in the concept map graph.
type ConceptMapNode struct {
	ID     string   `json:"id"`
	Label  string   `json:"label"`
	Type   NodeType `json:"type"`
	Detail string   `json:"detail,omitempty"`
}

// ConceptMapEdge connects two nodes.
type ConceptMapEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Label  string `json:"label"`
}

// ConceptMapData holds the full graph for rendering.
type ConceptMapData struct {
	Nodes []ConceptMapNode `json:"nodes"`
	Edges []ConceptMapEdge `json:"edges"`
	Topic string           `json:"topic"`
	Title string           `json:"title"` // winning idea title (or topic)
}

// BuildConceptMap extracts a concept map from a completed Discussion.
// It builds nodes for the winning idea, its pros/cons, runner-up ideas,
// key researcher findings, and implementer notes, then connects them with
// labelled edges.
func BuildConceptMap(d *models.Discussion) *ConceptMapData {
	if d == nil {
		return &ConceptMapData{Topic: "Unknown"}
	}

	data := &ConceptMapData{Topic: d.Topic, Title: d.Topic}
	nodeID := 0
	nextID := func() string {
		nodeID++
		return fmt.Sprintf("n%d", nodeID)
	}

	centerID := nextID()
	centerLabel := d.Topic
	centerDetail := ""

	if d.FinalIdea != nil {
		centerLabel = d.FinalIdea.Title
		centerDetail = d.FinalIdea.Description
		data.Title = d.FinalIdea.Title
	} else if len(d.Ideas) > 0 {
		best := d.Ideas[0]
		for _, idea := range d.Ideas {
			if idea.Score > best.Score {
				best = idea
			}
		}
		centerLabel = best.Title
		centerDetail = best.Description
		data.Title = best.Title
	}

	data.Nodes = append(data.Nodes, ConceptMapNode{
		ID:     centerID,
		Label:  truncate(centerLabel, 40),
		Type:   NodeCenter,
		Detail: centerDetail,
	})

	// Pros of the winning idea → green supporting nodes
	if d.FinalIdea != nil {
		for i, pro := range d.FinalIdea.Pros {
			if i >= 5 {
				break
			}
			id := nextID()
			data.Nodes = append(data.Nodes, ConceptMapNode{
				ID:    id,
				Label: truncate(pro, 45),
				Type:  NodePro,
			})
			data.Edges = append(data.Edges, ConceptMapEdge{Source: id, Target: centerID, Label: "supports"})
		}

		// Cons → red challenge nodes
		for i, con := range d.FinalIdea.Cons {
			if i >= 4 {
				break
			}
			id := nextID()
			data.Nodes = append(data.Nodes, ConceptMapNode{
				ID:    id,
				Label: truncate(con, 45),
				Type:  NodeCon,
			})
			data.Edges = append(data.Edges, ConceptMapEdge{Source: id, Target: centerID, Label: "challenges"})
		}
	}

	// Runner-up ideas → purple idea nodes
	added := 0
	for _, idea := range d.Ideas {
		if added >= 4 {
			break
		}
		if d.FinalIdea != nil && idea.Title == d.FinalIdea.Title {
			continue
		}
		id := nextID()
		data.Nodes = append(data.Nodes, ConceptMapNode{
			ID:     id,
			Label:  truncate(idea.Title, 40),
			Type:   NodeIdea,
			Detail: idea.Description,
		})
		data.Edges = append(data.Edges, ConceptMapEdge{Source: id, Target: centerID, Label: "alternative"})
		added++
	}

	// Researcher contributions → cyan research nodes (first sentence of each message)
	researchAdded := 0
	for _, msg := range d.Messages {
		if researchAdded >= 3 {
			break
		}
		if msg.From != "researcher" || msg.Type != "researcher" {
			continue
		}
		excerpt := firstSentence(msg.Content, 60)
		if excerpt == "" {
			continue
		}
		id := nextID()
		data.Nodes = append(data.Nodes, ConceptMapNode{
			ID:     id,
			Label:  excerpt,
			Type:   NodeResearch,
			Detail: truncate(msg.Content, 200),
		})
		data.Edges = append(data.Edges, ConceptMapEdge{Source: id, Target: centerID, Label: "grounded by"})
		researchAdded++
	}

	// Implementer contributions → orange implementation nodes
	implAdded := 0
	for _, msg := range d.Messages {
		if implAdded >= 2 {
			break
		}
		if msg.From != "implementer" || msg.Type != "implementer" {
			continue
		}
		excerpt := firstSentence(msg.Content, 60)
		if excerpt == "" {
			continue
		}
		id := nextID()
		data.Nodes = append(data.Nodes, ConceptMapNode{
			ID:     id,
			Label:  excerpt,
			Type:   NodeImplement,
			Detail: truncate(msg.Content, 200),
		})
		data.Edges = append(data.Edges, ConceptMapEdge{Source: id, Target: centerID, Label: "enabled by"})
		implAdded++
	}

	return data
}

// RenderConceptMapHTML returns a self-contained HTML section containing an
// interactive D3.js v7 force-directed concept map.  It is designed to be
// injected into the idea sheet just before </body>.
func RenderConceptMapHTML(data *ConceptMapData) string {
	nodesJSON, _ := json.Marshal(data.Nodes)
	edgesJSON, _ := json.Marshal(data.Edges)
	titleJSON, _ := json.Marshal(data.Title)
	topicJSON, _ := json.Marshal(data.Topic)

	return fmt.Sprintf(`
<section id="concept-map" style="font-family:sans-serif;padding:40px 20px;background:#0d1117;color:#e6edf3;page-break-before:always;">
  <h2 style="text-align:center;font-size:1.6rem;margin-bottom:4px;color:#fff;">🗺️ Concept Map</h2>
  <p style="text-align:center;color:#8b949e;font-size:0.85rem;margin-bottom:24px;">Topic: %s</p>
  <div id="cm-wrap" style="position:relative;width:100%%;max-width:960px;margin:0 auto;background:#161b22;border-radius:16px;border:1px solid rgba(255,255,255,0.08);overflow:hidden;">
    <svg id="cm-svg" style="display:block;width:100%%;height:580px;"></svg>
    <div id="cm-tooltip" style="position:absolute;display:none;background:rgba(13,17,23,0.96);border:1px solid rgba(255,255,255,0.12);border-radius:8px;padding:10px 14px;font-size:0.78rem;max-width:260px;pointer-events:none;line-height:1.5;color:#e6edf3;z-index:99;"></div>
  </div>
  <div id="cm-legend" style="display:flex;flex-wrap:wrap;justify-content:center;gap:12px 24px;margin-top:20px;font-size:0.75rem;color:#8b949e;"></div>
  <script src="https://cdn.jsdelivr.net/npm/d3@7/dist/d3.min.js"></script>
  <script>
  (function(){
    var nodes = %s;
    var edges = %s;
    var title = %s;

    var palette = {
      center:    {fill:"#FF6B6B", r:44,  stroke:"#ff9999"},
      pro:       {fill:"#51E898", r:20,  stroke:"#a0f4c0"},
      con:       {fill:"#FFD93D", r:18,  stroke:"#ffe080"},
      idea:      {fill:"#7B68EE", r:26,  stroke:"#a89ff5"},
      research:  {fill:"#00D4FF", r:18,  stroke:"#66e3ff"},
      implement: {fill:"#FF8C42", r:22,  stroke:"#ffb07a"},
    };

    var legendLabels = {
      center:"Winning Idea", pro:"Strength", con:"Challenge",
      idea:"Alternative", research:"Research", implement:"Implementation"
    };

    // Legend
    var leg = document.getElementById('cm-legend');
    Object.keys(palette).forEach(function(t){
      if (!nodes.find(function(n){return n.type===t;})) return;
      var el = document.createElement('span');
      el.style.display='flex'; el.style.alignItems='center'; el.style.gap='6px';
      el.innerHTML='<svg width="14" height="14"><circle cx="7" cy="7" r="6" fill="'+palette[t].fill+'"/></svg>'+legendLabels[t];
      leg.appendChild(el);
    });

    var wrap = document.getElementById('cm-wrap');
    var W = wrap.offsetWidth || 960, H = 580;
    var svg = d3.select('#cm-svg').attr('viewBox','0 0 '+W+' '+H);
    var tooltip = document.getElementById('cm-tooltip');

    var g = svg.append('g');

    // Zoom + pan
    svg.call(d3.zoom().scaleExtent([0.3,3]).on('zoom',function(e){g.attr('transform',e.transform);}));

    // Arrow marker
    svg.append('defs').append('marker')
      .attr('id','arr').attr('viewBox','0 -4 8 8').attr('refX',8).attr('refY',0)
      .attr('markerWidth',6).attr('markerHeight',6).attr('orient','auto')
      .append('path').attr('d','M0,-4L8,0L0,4').attr('fill','rgba(255,255,255,0.25)');

    var linkSel = g.append('g').selectAll('line')
      .data(edges).join('line')
      .attr('stroke','rgba(255,255,255,0.18)')
      .attr('stroke-width',1.5)
      .attr('marker-end','url(#arr)');

    var linkLabelSel = g.append('g').selectAll('text')
      .data(edges).join('text')
      .attr('fill','rgba(255,255,255,0.35)')
      .attr('font-size','9px')
      .attr('text-anchor','middle')
      .text(function(d){return d.label;});

    var nodeSel = g.append('g').selectAll('g')
      .data(nodes).join('g')
      .style('cursor','pointer')
      .call(d3.drag()
        .on('start',function(e,d){if(!e.active)sim.alphaTarget(0.3).restart();d.fx=d.x;d.fy=d.y;})
        .on('drag', function(e,d){d.fx=e.x;d.fy=e.y;})
        .on('end',  function(e,d){if(!e.active)sim.alphaTarget(0);d.fx=null;d.fy=null;}));

    nodeSel.append('circle')
      .attr('r',function(d){return palette[d.type]?palette[d.type].r:18;})
      .attr('fill',function(d){return palette[d.type]?palette[d.type].fill:'#888';})
      .attr('stroke',function(d){return palette[d.type]?palette[d.type].stroke:'#aaa';})
      .attr('stroke-width',2);

    nodeSel.append('text')
      .attr('text-anchor','middle')
      .attr('dominant-baseline','middle')
      .attr('fill','#000')
      .attr('font-weight','700')
      .attr('font-size',function(d){return d.type==='center'?'11px':'9px';})
      .attr('pointer-events','none')
      .each(function(d){
        var el=d3.select(this);
        var r=palette[d.type]?palette[d.type].r:18;
        var maxW=r*1.7;
        var words=d.label.split(' ');
        var line='', lines=[];
        words.forEach(function(w){
          var test=line?line+' '+w:w;
          if(test.length*6>maxW&&line){lines.push(line);line=w;}
          else{line=test;}
        });
        lines.push(line);
        var dy=-(lines.length-1)*5.5;
        lines.forEach(function(l,i){
          el.append('tspan').attr('x',0).attr('dy',i===0?dy+'px':'11px').text(l);
        });
      });

    // Tooltip
    nodeSel.on('mouseenter',function(e,d){
      if(!d.detail&&!d.label)return;
      tooltip.style.display='block';
      tooltip.innerHTML='<strong style="color:'+
        (palette[d.type]?palette[d.type].fill:'#fff')+'">'+d.label+'</strong>'+
        (d.detail?'<br><span style="color:#8b949e">'+d.detail+'</span>':'');
    }).on('mousemove',function(e){
      var b=wrap.getBoundingClientRect();
      var x=e.clientX-b.left+12, y=e.clientY-b.top-10;
      if(x+270>W)x=x-290;
      tooltip.style.left=x+'px';
      tooltip.style.top=y+'px';
    }).on('mouseleave',function(){tooltip.style.display='none';});

    var sim = d3.forceSimulation(nodes)
      .force('link', d3.forceLink(edges).id(function(d){return d.id;}).distance(function(d){
        var s=nodes.find(function(n){return n.id===d.source.id||n.id===d.source;});
        return s&&s.type==='center'?130:90;
      }).strength(0.6))
      .force('charge', d3.forceManyBody().strength(-320))
      .force('center', d3.forceCenter(W/2, H/2))
      .force('collide', d3.forceCollide().radius(function(d){return (palette[d.type]?palette[d.type].r:18)+12;}))
      .on('tick', function(){
        linkSel
          .attr('x1',function(d){return d.source.x;}).attr('y1',function(d){return d.source.y;})
          .attr('x2',function(d){return d.target.x;}).attr('y2',function(d){return d.target.y;});
        linkLabelSel
          .attr('x',function(d){return (d.source.x+d.target.x)/2;})
          .attr('y',function(d){return (d.source.y+d.target.y)/2-4;});
        nodeSel.attr('transform',function(d){return 'translate('+d.x+','+d.y+')';});
      });
  })();
  </script>
</section>
`, topicJSON, nodesJSON, edgesJSON, titleJSON)
}

// InjectIntoHTML appends the concept map section into an HTML document.
// It inserts before </body> if present, otherwise appends to the end.
func InjectIntoHTML(html string, section string) string {
	lower := strings.ToLower(html)
	if idx := strings.LastIndex(lower, "</body>"); idx >= 0 {
		return html[:idx] + section + html[idx:]
	}
	return html + section
}

// truncate shortens s to at most n runes, appending "…" if cut.
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-1]) + "…"
}

// firstSentence returns the first sentence of s, capped at maxLen runes.
// Sentence endings detected: ". ", "! ", "? ", or a newline.
func firstSentence(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	for i, r := range s {
		if (r == '.' || r == '!' || r == '?') && i+1 < len(s) {
			return truncate(s[:i+1], maxLen)
		}
		if r == '\n' {
			return truncate(s[:i], maxLen)
		}
	}
	return truncate(s, maxLen)
}
