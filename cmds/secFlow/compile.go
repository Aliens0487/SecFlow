package main

import (
	"log"

	"github.com/Aliens0487/secFlow/pkg/zkp"
	"github.com/spf13/cobra"
)

var compileCommand = &cobra.Command{
	Use:   "compile <model>",
	Short: "Compile a BPMN file into a zero-knowledge circuit",
	Args:  cobra.ExactArgs(1),
	Run:   compileCMDExecute,
}

func init() {
	compileCommand.PersistentFlags().StringP("output", "o", "circuit.r1cs", "Output file for the compiled circuit")
	rootCMD.AddCommand(compileCommand)
}

func compileCMDExecute(cmd *cobra.Command, args []string) {
	modelPath := args[0]
	outputFlag, _ := cmd.Flags().GetString("output")

	secFlow, err := zkp.NewSecFlowProgram(modelPath)
	if err != nil {
		log.Fatalln("Failed to create secFlow program:", err)
	}

	err = secFlow.Compile(outputFlag)
	if err != nil {
		log.Fatalln("Failed to compile secFlow program:", err)
	}
}
