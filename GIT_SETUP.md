# Git Setup Guide

Your repository is initialized and ready! Here's how to connect it to GitHub and set up your workflow.

## Current Status

✅ Git repository initialized
✅ Initial commit created (36 files, 6941+ lines)
✅ Development documentation added
✅ GitHub templates configured
✅ Proper .gitignore in place

## Next Steps

### 1. Create GitHub Repository

Go to [GitHub](https://github.com/new) and create a new repository:

**Repository Settings:**
- Name: `ai-agent-team`
- Description: "Multi-agent collaborative ideation system with configurable teams, multi-round discussions, and comprehensive report generation"
- Visibility: Public or Private (your choice)
- **DO NOT** initialize with README, .gitignore, or license (we already have these)

### 2. Connect to Remote

After creating the GitHub repo, run these commands:

```bash
# Add your GitHub repo as remote (replace with your URL)
git remote add origin https://github.com/yourusername/ai-agent-team.git

# Or if using SSH:
git remote add origin git@github.com:yourusername/ai-agent-team.git

# Verify remote is added
git remote -v

# Push to GitHub
git branch -M main
git push -u origin main
```

### 3. Verify on GitHub

Check that everything pushed correctly:
- ✅ All files visible
- ✅ README.md displays on homepage
- ✅ Issues tab shows templates
- ✅ Actions tab shows workflow

## Repository Configuration

### Branch Protection (Recommended)

Protect your main branch:

1. Go to Settings → Branches
2. Add rule for `main` branch
3. Enable:
   - ✅ Require pull request before merging
   - ✅ Require status checks to pass (once Actions are set up)
   - ✅ Require linear history
   - ✅ Include administrators

### GitHub Actions

The `.github/workflows/test.yml` workflow will:
- ✅ Run on push to main/develop
- ✅ Run on pull requests
- ✅ Build all binaries
- ✅ Run `go vet`
- ✅ Check formatting

First run will happen automatically after push.

### Topics/Tags (Optional)

Add topics to help people discover your repo:
- `ai`
- `agents`
- `collaboration`
- `ideation`
- `multi-agent`
- `anthropic-claude`
- `golang`
- `tui`
- `bubbletea`

Settings → About (right sidebar) → Topics

## Development Workflow

### Creating a Feature

```bash
# Create feature branch
git checkout -b feature/your-feature-name

# Make changes
# ... edit files ...

# Check changes
make check

# Commit
git add .
git commit -m "feat: add your feature"

# Push
git push origin feature/your-feature-name
```

### Creating a Pull Request

1. Push your branch to GitHub
2. Go to repository on GitHub
3. Click "Compare & pull request"
4. Fill out the PR template
5. Request review (if working with others)
6. Merge after approval

### Keeping Your Fork Updated

If you forked from another repo:

```bash
# Add upstream remote
git remote add upstream https://github.com/original/ai-agent-team.git

# Fetch updates
git fetch upstream

# Merge into your main
git checkout main
git merge upstream/main
git push origin main
```

## Git Best Practices

### Commit Messages

Follow Conventional Commits:

```bash
feat: add new researcher agent
fix: correct progress bar in TUI
docs: update README with examples
refactor: simplify orchestrator logic
test: add test for agent communication
chore: update dependencies
```

### Branch Naming

```bash
feature/add-memory-system
bugfix/fix-tui-crash
docs/improve-quickstart
refactor/simplify-api-client
```

### Before Committing

```bash
# Format code
make fmt

# Run checks
make check

# Review changes
git diff

# Stage selectively
git add -p
```

## Useful Git Commands

### Check Status
```bash
git status
git log --oneline -10
git diff
```

### Undo Changes
```bash
# Undo unstaged changes
git checkout -- filename

# Undo last commit (keep changes)
git reset --soft HEAD~1

# Undo last commit (discard changes)
git reset --hard HEAD~1
```

### Working with Branches
```bash
# List branches
git branch -a

# Switch branches
git checkout branch-name

# Delete local branch
git branch -d branch-name

# Delete remote branch
git push origin --delete branch-name
```

### Viewing History
```bash
# Pretty log
git log --oneline --graph --all

# See what changed in a commit
git show commit-hash

# See file history
git log -p filename
```

## Makefile Commands

Quick reference for development:

```bash
make build        # Build all binaries
make clean        # Remove build artifacts
make fmt          # Format code
make vet          # Run go vet
make check        # fmt + vet + build
make deps         # Install dependencies
make run-tui      # Run TUI
make run-cli      # Run CLI v2
make run-server   # Run web server
make help         # Show all commands
```

## Files Already Configured

✅ `.gitignore` - Excludes binaries, generated files, etc.
✅ `CONTRIBUTING.md` - Contribution guidelines
✅ `DEVELOPMENT.md` - Architecture and development guide
✅ `Makefile` - Build automation
✅ `.github/workflows/test.yml` - CI pipeline
✅ `.github/ISSUE_TEMPLATE/` - Issue templates
✅ `.github/pull_request_template.md` - PR template

## Initial Commits

Your repository has 2 commits:

1. **feat: initial release** - Complete codebase (36 files)
2. **docs: add development docs** - GitHub templates and dev guides

## Quick Start for Contributors

Anyone cloning your repo can get started with:

```bash
git clone https://github.com/yourusername/ai-agent-team.git
cd ai-agent-team
make deps
make build
export ANTHROPIC_API_KEY="their-key"
make run-tui
```

## Security Notes

### API Keys
- ✅ `.gitignore` includes `.env` files
- ✅ Never commit API keys
- ✅ Use environment variables
- ✅ Document in README that users need their own key

### Secrets in GitHub Actions
If you need secrets for CI:
1. Go to Settings → Secrets and variables → Actions
2. Add `ANTHROPIC_API_KEY` (if testing with API)
3. Reference in workflow: `${{ secrets.ANTHROPIC_API_KEY }}`

## Next Steps

After pushing to GitHub:

1. ✅ Add topics/tags
2. ✅ Add repository description
3. ✅ Enable Discussions (if you want community input)
4. ✅ Add a LICENSE file (MIT recommended)
5. ✅ Consider adding a SECURITY.md for security policy
6. ✅ Star your own repo (why not!)

## Questions?

- Check `CONTRIBUTING.md` for contribution workflow
- Check `DEVELOPMENT.md` for architecture details
- Check `README_V2.md` for feature documentation

---

**Your repository is production-ready! 🚀**

All documentation, best practices, and tooling are in place.
Just connect to GitHub and start collaborating!
