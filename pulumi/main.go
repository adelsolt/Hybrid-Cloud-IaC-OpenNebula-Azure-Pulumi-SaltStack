package main

import (
        "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
        pulumi.Run(func(ctx *pulumi.Context) error {
                if err := buildOpenNebula(ctx); err != nil {
                        return err
                }
                return buildAzure(ctx)

        })
}