cfg := config.New(ctx, "")

raw, _ := os.ReadFile("bootstrap/minion.sh.tmpl")
script := strings.NewReplacer(
    "__MASTER_ADDR__", masterAddrInternal,
    "__MINION_ID__",   m.Name,
    "__ROLE__",        m.Role,
).Replace(string(raw))
b64 := base64.StdEncoding.EncodeToString([]byte(script))

Context: pulumi.Map{
    "NETWORK":             pulumi.String("YES"),
    "SSH_PUBLIC_KEY":      pulumi.String(cfg.Require("sshPublicKey")),
    "START_SCRIPT_BASE64": pulumi.String(b64),
},