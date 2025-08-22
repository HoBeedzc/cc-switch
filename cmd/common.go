package cmd

import (
	"fmt"

	"cc-switch/internal/config"
	"cc-switch/internal/ui"
)

// handleCurrentConfigError 处理获取当前配置时的特殊错误
func handleCurrentConfigError(err error, uiProvider ui.UIProvider) error {
	switch e := err.(type) {
	case *config.EmptyModeError:
		uiProvider.ShowWarning(e.Message)
		for _, suggestion := range e.Suggestions {
			fmt.Printf("  %s\n", suggestion)
		}
		return nil

	case *config.NoCurrentProfileError:
		uiProvider.ShowError(fmt.Errorf(e.Message))
		fmt.Println("Available options:")
		for _, suggestion := range e.Suggestions {
			fmt.Printf("  %s\n", suggestion)
		}
		return err

	case *config.ProfileMissingError:
		uiProvider.ShowError(fmt.Errorf(e.Message))
		fmt.Println("The current configuration file has been deleted or moved.")
		fmt.Println("Available options:")
		for _, suggestion := range e.Suggestions {
			fmt.Printf("  %s\n", suggestion)
		}
		return err

	default:
		return err
	}
}
