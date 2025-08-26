package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/iurnickita/gophkeeper1/internal/client/service"
	"github.com/spf13/cobra"
)

// Execute инициализация CLI интерфейса
func Execute(service service.Service) {
	handler := cliHandler{service: service}

	// root
	var rootCmd = &cobra.Command{
		Use:   "gophkpr",
		Short: "GophKeeper менеджер паролей",
		Long:  "GophKeeper менеджер паролей",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}

	// Register
	var registerCmd = &cobra.Command{
		Use:     "rg",
		Aliases: []string{"register"},
		Short:   "Register: rg <login> <password>",
		Long:    "Register регистрирует нового пользователя. Формат ввода: rg <login> <password>",
		Args:    cobra.ExactArgs(2),
		Run:     handler.register,
	}
	rootCmd.AddCommand(registerCmd)

	// Login
	var loginCmd = &cobra.Command{
		Use:     "lg",
		Aliases: []string{"login"},
		Short:   "Login: lg <login> <password>",
		Long:    "Login производит вход на устройстве. Формат ввода: lg <login> <password>",
		Args:    cobra.ExactArgs(2),
		Run:     handler.login,
	}
	rootCmd.AddCommand(loginCmd)

	// List
	var listCmd = &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List",
		Long:    "List возвращает список имен доступных данных",
		Run:     handler.list,
	}
	rootCmd.AddCommand(listCmd)

	// Read
	var readTarget string
	var readCmd = &cobra.Command{
		Use:     "rd",
		Aliases: []string{"read"},
		Short:   "Read: rd <unitname>",
		Long:    "Read возвращает единицу данных по имени. Формат ввода: rd <unitname>",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			handler.read(args, readTarget)
		},
	}
	readCmd.Flags().StringVarP(&readTarget, "target", "t", "", "Target directory to result output")
	rootCmd.AddCommand(readCmd)

	// Write
	var writeSource string
	var writeCmd = &cobra.Command{
		Use:     "wr",
		Aliases: []string{"write"},
		Short:   "Write: wr <unitname> <type> <data>",
		Long: `Write сохраняет единицу данных. Формат ввода: wr <unitname> <type> <data>
	Доступные <type>:
	1 - login. <data>: "login password"
	2 - text. <data>: "text"
	3 - binary. <data>: _
	4 - card. <data>: "number yearmonth name surname cvv"`,
		Args: cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			handler.write(args, writeSource)
		},
	}
	writeCmd.Flags().StringVarP(&writeSource, "source", "s", "", "Source directory to read from")
	rootCmd.AddCommand(writeCmd)

	// Delete
	var deleteCmd = &cobra.Command{
		Use:     "dl",
		Aliases: []string{"delete"},
		Short:   "Delete: dl <unitname>",
		Long:    "Delete удаляет единицу данных. Формат ввода: dl <unitname>",
		Args:    cobra.ExactArgs(1),
		Run:     handler.delete,
	}
	rootCmd.AddCommand(deleteCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка выполнения GophKeeper '%s'\n", err)
		os.Exit(1)
	}
}

// cliHandler обработчик CLI команд
type cliHandler struct {
	service service.Service
}

// Register
func (h cliHandler) register(cmd *cobra.Command, args []string) {
	err := h.service.Register(args[0], args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	fmt.Fprintln(os.Stdout, "OK")
}

// Login
func (h cliHandler) login(cmd *cobra.Command, args []string) {
	err := h.service.Login(args[0], args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	fmt.Fprintln(os.Stdout, "OK")
}

// List
func (h cliHandler) list(cmd *cobra.Command, args []string) {
	list, err := h.service.List()
	if err != nil {
		switch err {
		case service.ErrOffline:
			// Офлайн. Выводим результат с предупреждением
			fmt.Fprintln(os.Stdout, err.Error())
		default:
			// Ошибка
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}
	}
	fmt.Fprintln(os.Stdout, list)
}

// Read
func (h cliHandler) read(args []string, target string) {
	unitValue, err := h.service.Read(args[0], target)
	if err != nil {
		switch err {
		case service.ErrOffline:
			// Офлайн. Выводим результат с предупреждением
			fmt.Fprintln(os.Stdout, err.Error())
		default:
			// Ошибка
			fmt.Fprintln(os.Stderr, err.Error())
			return
		}
	}

	fmt.Fprintln(os.Stdout, unitValue)
}

// Write
func (h cliHandler) write(args []string, source string) {
	unitName := args[0]
	unitType, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	unitValue := args[2]

	if err := h.service.Write(unitName, unitType, unitValue, source); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	fmt.Fprintln(os.Stdout, "OK")
}

// Delete
func (h cliHandler) delete(cmd *cobra.Command, args []string) {
	err := h.service.Delete(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	fmt.Fprintln(os.Stdout, "OK")
}
