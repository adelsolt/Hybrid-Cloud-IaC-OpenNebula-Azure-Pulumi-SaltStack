package main

import (
        "encoding/base64"
        "fmt"
        "os"
        "strings"
        "github.com/pulumi/pulumi-terraform-provider/sdks/go/opennebula/opennebula"
        "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
        "github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// one VM's spec
type minion struct {
        Name string
        Role string
        CPU  float64
        RAM  int // MB
}

func buildOpenNebula(ctx *pulumi.Context) error {
        cfg := config.New(ctx, "")    // sshPublicKey, masterAddrInternal
        one := config.New(ctx, "one") // endpoint, username, password

        // My own provider (ofc OpenNebula here) credentials from the 'one' config I injected earlier
        prov, err := opennebula.NewProvider(ctx, "one", &opennebula.ProviderArgs{
                Endpoint: pulumi.String(one.Require("endpoint")),
                Username: pulumi.String(one.Require("username")),
                Password: one.RequireSecret("password"),
        })
        if err != nil {
                return err
        }

        // this would do read the bootstrap template once
        raw, err := os.ReadFile("bootstrap/minion.sh.tmpl")
        if err != nil {
                return err
        }

        sshKey := cfg.Require("sshPublicKey")
        masterAddr := cfg.Require("masterAddrInternal")

        templateID := 56

        fleet := []minion{
                {"db-01", "db", 1, 2048},
                {"app-01", "app", 2, 2048},
        }

        for _, m := range fleet {
                script := strings.NewReplacer(
                        "__MASTER_ADDR__", masterAddr,
                        "__MINION_ID__", m.Name,
                        "__ROLE__", m.Role,
                ).Replace(string(raw))
                b64 := base64.StdEncoding.EncodeToString([]byte(script))


                vm, err := opennebula.NewVirtualMachine(ctx, m.Name, &opennebula.VirtualMachineArgs{
                        Cpu:        pulumi.Float64(m.CPU),
                        Vcpu:       pulumi.Float64(m.CPU),
                        Memory:     pulumi.Float64(m.RAM),
                        TemplateId: pulumi.Float64(templateID),
                        Context: pulumi.StringMap{
                                "NETWORK":             pulumi.String("YES"),
                                "SSH_PUBLIC_KEY":      pulumi.String(sshKey),
                                "START_SCRIPT_BASE64": pulumi.String(b64),
                        },
                }, pulumi.Provider(prov))
                if err != nil {
                        return err
                }

                ctx.Export(fmt.Sprintf("%s-id", m.Name), vm.ID())
        }

        return nil
}