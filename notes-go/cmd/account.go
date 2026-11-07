package cmd

import (
	plumcliutilsargs "github.com/cyril-fl/plum/cli/utils/args"
	"github.com/spf13/cobra"
	cmdaccount "main.go/app/cli/cmd-account"
)

var cmdLogin = &cobra.Command{
	Use:     cmd.Login.Use,
	Short:   cmd.Login.Short,
	Long:    cmd.Login.Long,
	Example: cmd.Login.Example,
	Args:    cobra.NoArgs,
	RunE:    cmdaccount.Login,
}

var cmdLogout = &cobra.Command{
	Use:   cmd.Logout.Use,
	Short: cmd.Logout.Short,
	Args:  cobra.NoArgs,
	RunE:  cmdaccount.Logout,
}

var cmdUser = &cobra.Command{
	Use:   cmd.User.Use,
	Short: cmd.User.Short,
}

var cmdUserAdd = &cobra.Command{
	Use:     cmd.UserAdd.Use,
	Short:   cmd.UserAdd.Short,
	Long:    cmd.UserAdd.Long,
	Example: cmd.UserAdd.Example,
	Args:    cobra.ExactArgs(1),
	RunE:    cmdaccount.AddUser,
}

var cmdResetPassword = &cobra.Command{
	Use:     cmd.ResetPassword.Use,
	Short:   cmd.ResetPassword.Short,
	Long:    cmd.ResetPassword.Long,
	Example: cmd.ResetPassword.Example,
	Args:    cobra.ExactArgs(1),
	RunE:    cmdaccount.ResetPassword,
}

var cmdRotateSecret = &cobra.Command{
	Use:     cmd.RotateSecret.Use,
	Short:   cmd.RotateSecret.Short,
	Long:    cmd.RotateSecret.Long,
	Example: cmd.RotateSecret.Example,
	Args:    cobra.NoArgs,
	RunE:    cmdaccount.RotateSecret,
}

var cmdAccount = &cobra.Command{
	Use:   cmd.Account.Use,
	Short: cmd.Account.Short,
}

var cmdAccountPurge = &cobra.Command{
	Use:     cmd.AccountPurge.Use,
	Short:   cmd.AccountPurge.Short,
	Long:    cmd.AccountPurge.Long,
	Example: cmd.AccountPurge.Example,
	Args:    cobra.NoArgs,
	RunE:    cmdaccount.Purge,
}

// noAccount is what runs without a logged-in account: logging in, the admin's commands, and the API, which logs
// each request in itself.
var noAccount = map[*cobra.Command]bool{cmdUser: true, cmdLogin: true, cmdLogout: true, cmdUserAdd: true, cmdResetPassword: true, cmdRotateSecret: true, cmdServe: true}

func init() {
	plumcliutilsargs.Makeflags(cmdUserAdd, cmd.UserAdd.Flags)
	plumcliutilsargs.Makeflags(cmdResetPassword, cmd.ResetPassword.Flags)

	cmdUser.AddCommand(cmdUserAdd)
	cmdAccount.AddCommand(cmdAccountPurge)
	cmdRoot.AddCommand(cmdLogin, cmdLogout, cmdUser, cmdResetPassword, cmdRotateSecret, cmdAccount)
}
