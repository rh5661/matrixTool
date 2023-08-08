/*
Copyright © 2023 ROBERT HUANG
*/
package cmd

import (
	"fmt"
	"github.com/rh5661/matrixTool/pkg/dbModify"
	"github.com/rh5661/matrixTool/pkg/excel"

	"github.com/spf13/cobra"
)

// loadCmd represents the load command
var (
	filePath string

	loadCmd = &cobra.Command{
		Use:   "load",
		Short: "Loads matrix excel file into database",
		Long: `Excel filepath inserted will be parsed and injected into database
Please specify the absolute filepath from the root
Example usage:
matrixTool load "C:\Users\Robert\Downloads\Daily Matrix Price For All Term.xlsx"`,
		Run: func(cmd *cobra.Command, args []string) {
			if filePath == "" && len(args) != 0 {
				filePath = args[0]
				fmt.Println("The entered filePath is: " + filePath + "\n")
			} else {
				fmt.Println("Please specify a filepath. Run 'matrixTool load --help' for more information.")
				return
			}

			dbModify.SetFilePath(filePath)
			excel.ReadExcelFile(filePath)
			fmt.Println("\nFinished loading")
		},
	}
)

func init() {
	rootCmd.AddCommand(loadCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	loadCmd.PersistentFlags().StringVarP(&filePath, "filePath", "", "", "[Required] Set filepath of matrix pricing Excel file to be parsed")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loadCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
