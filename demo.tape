# kubectl-generate demo
# Requires: kind cluster with CRDs installed

Output demo.gif
Output demo.mp4

Set Shell zsh
Set FontSize 15
Set Width 1000
Set Height 600
Set Theme "Catppuccin Mocha"
Set Padding 20
Set TypingSpeed 40ms

# Title
Type "# kubectl-generate: schema-driven YAML from your cluster"
Enter
Sleep 1.5s

# Show available types
Type "kubectl generate --list | head -20"
Enter
Sleep 3s

# Generate a Deployment
Type "kubectl generate Deployment --name=web --image=myapp:v2 --replicas=3"
Enter
Sleep 4s

# Fuzzy matching on typo
Type "kubectl generate Deploymnet"
Enter
Sleep 3s

# Generate a CronJob with --set override
Type "kubectl generate CronJob --name=backup --set spec.schedule='*/5 * * * *'"
Enter
Sleep 4s

# Generate a CRD (Gateway API)
Type "kubectl generate HTTPRoute --name=api"
Enter
Sleep 4s

# Pipe to kubectl apply --dry-run
Type "kubectl generate Service --name=web | kubectl apply --dry-run=server -f -"
Enter
Sleep 3s

# Done
Type ""
Sleep 2s
