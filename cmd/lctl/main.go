package main
import (
	"context"
	"github.com/example/lctl/cmd/lctl/cli" // Adjusted
)
func main() {
	ctx := context.Background()
	cli.Execute(ctx)
}
