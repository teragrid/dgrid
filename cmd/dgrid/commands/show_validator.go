package commands

import (
	"fmt"
	"github.com/spf13/cobra"

	validator "github.com/teragrid/dgrid/core/types/validator"
)

// ShowValidatorCmd adds capabilities for showing the validator info.
var ShowValidatorCmd = &cobra.Command{
	Use:   "show_validator",
	Short: "Show this node's validator info",
	Run:   showValidator,
}

func showValidator(cmd *cobra.Command, args []string) {
	validator := validator.LoadOrGenFilePV(config.ValidatorFile())
	pubKeyJSONBytes, _ := cdc.MarshalJSON(validator.GetPubKey())
	fmt.Println(string(pubKeyJSONBytes))
}
