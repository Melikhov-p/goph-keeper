// main package of client side of app.
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	pb "github.com/Melikhov-p/goph-keeper/internal/api/gen"
	"golang.org/x/term"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const (
	cardNumberSafeLen = 4
	serverAddress     = "localhost:50051"
)

var (
	userClient   pb.UserServiceClient
	secretClient pb.SecretServiceClient
	token        string
)

func main() {
	var (
		err  error
		conn *grpc.ClientConn
	)
	conn, err = grpc.NewClient(serverAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Printf("did not connect: %v\n", err)
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	userClient = pb.NewUserServiceClient(conn)
	secretClient = pb.NewSecretServiceClient(conn)

	showMainMenu()
}

func showMainMenu() {
	for {
		if token == "" {
			fmt.Println("\nGophKeeper Client")
			fmt.Println("1. Register")
			fmt.Println("2. Login")
			fmt.Println("3. Exit")
		} else {
			fmt.Println("\n4. Create secret")
			fmt.Println("5. Get secrets")
			fmt.Println("6. Logout")
		}

		fmt.Print("Select an option: ")
		reader := bufio.NewReader(os.Stdin)
		option, _ := reader.ReadString('\n')
		option = strings.TrimSpace(option)

		switch option {
		case "1":
			registerUser()
		case "2":
			loginUser()
		case "3":
			fmt.Println("Goodbye!")
			return
		case "4":
			if token != "" {
				updateUser()
			} else {
				fmt.Println("Invalid option")
			}
		case "5":
			if token != "" {
				createSecret()
			} else {
				fmt.Println("Invalid option")
			}
		case "6":
			if token != "" {
				getSecrets()
			} else {
				fmt.Println("Invalid option")
			}
		case "7":
			if token != "" {
				token = ""
				fmt.Println("Logged out successfully")
			} else {
				fmt.Println("Invalid option")
			}
		default:
			fmt.Println("Invalid option")
		}
	}
}

func registerUser() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter login: ")
	login, _ := reader.ReadString('\n')
	login = strings.TrimSpace(login)

	fmt.Print("Enter password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	req := &pb.RegisterUserRequest{
		Login:    login,
		Password: password,
	}

	res, err := userClient.Register(context.Background(), req)
	if err != nil {
		fmt.Printf("Registration failed: %v\n", err)
		return
	}

	fmt.Printf("Registered successfully. User ID: %d\n", res.GetUser().GetId())
}

func loginUser() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter login: ")
	login, _ := reader.ReadString('\n')
	login = strings.TrimSpace(login)

	fmt.Print("Enter password: ")
	// Читаем пароль с помощью term.ReadPassword
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		fmt.Printf("\nFailed to read password: %v\n", err)
		return
	}
	password := strings.TrimSpace(string(bytePassword))
	fmt.Println() // Добавляем перевод строки после ввода пароля

	// Проверяем, что пароль не пустой
	if password == "" {
		fmt.Println("Error: Password cannot be empty")
		return
	}

	req := &pb.LoginUserRequest{
		Login:    login,
		Password: password, // Уже обрезаны пробелы
	}

	var header metadata.MD
	res, err := userClient.Login(
		context.Background(),
		req,
		grpc.Header(&header),
	)
	if err != nil {
		fmt.Printf("\nLogin failed: %v\n", err)
		return
	}

	// Получаем токен из заголовков
	if authHeaders := header.Get("authorization"); len(authHeaders) > 0 {
		token = authHeaders[0]
		fmt.Printf("\nLogged in successfully. Welcome, %s!\n", res.GetUser().GetLogin())
	} else {
		fmt.Println("\nWarning: Server didn't return authorization token")
	}
}

func updateUser() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter new login (leave empty to keep current): ")
	newLogin, _ := reader.ReadString('\n')
	newLogin = strings.TrimSpace(newLogin)

	fmt.Print("Enter old password: ")
	oldPassword, _ := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()

	fmt.Print("Enter new password (leave empty to keep current): ")
	newPassword, _ := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()

	req := &pb.UpdateUserRequest{
		NewLogin:    newLogin,
		OldPassword: strings.TrimSpace(string(oldPassword)),
		NewPassword: strings.TrimSpace(string(newPassword)),
	}

	ctx := withToken(context.Background())
	_, err := userClient.Update(ctx, req)
	if err != nil {
		fmt.Printf("Update failed: %v\n", err)
		return
	}

	fmt.Println("Credentials updated successfully")
}

func createSecret() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\nSelect secret type:")
	fmt.Println("1. Password")
	fmt.Println("2. Credit Card")
	fmt.Println("3. Binary Data")
	fmt.Print("Your choice: ")
	typeChoice, _ := reader.ReadString('\n')
	typeChoice = strings.TrimSpace(typeChoice)

	var secretType pb.SecretType
	var secretData interface{}

	fmt.Print("Enter secret name: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	switch typeChoice {
	case "1":
		secretType = pb.SecretType_SECRET_TYPE_PASSWORD
		fmt.Print("Enter username: ")
		username, _ := reader.ReadString('\n')
		username = strings.TrimSpace(username)

		fmt.Print("Enter password: ")
		password, _ := reader.ReadString('\n')
		password = strings.TrimSpace(password)

		fmt.Print("Enter URL (optional): ")
		url, _ := reader.ReadString('\n')
		url = strings.TrimSpace(url)

		fmt.Print("Enter notes (optional): ")
		notes, _ := reader.ReadString('\n')
		notes = strings.TrimSpace(notes)

		secretData = &pb.PasswordData{
			Username: username,
			Password: password,
			Url:      url,
			Notes:    &notes,
		}
	case "2":
		secretType = pb.SecretType_SECRET_TYPE_CARD
		fmt.Print("Enter card owner: ")
		owner, _ := reader.ReadString('\n')
		owner = strings.TrimSpace(owner)

		fmt.Print("Enter card number: ")
		number, _ := reader.ReadString('\n')
		number = strings.TrimSpace(number)

		fmt.Print("Enter CVV: ")
		cvv, _ := reader.ReadString('\n')
		cvv = strings.TrimSpace(cvv)

		fmt.Print("Enter expire date: ")
		expireDate, _ := reader.ReadString('\n')
		expireDate = strings.TrimSpace(expireDate)

		fmt.Print("Enter notes (optional): ")
		notes, _ := reader.ReadString('\n')
		notes = strings.TrimSpace(notes)

		secretData = &pb.CardData{
			Owner:      owner,
			Number:     number,
			CVV:        cvv,
			ExpireDate: expireDate,
			Notes:      &notes,
		}
	case "3":
		secretType = pb.SecretType_SECRET_TYPE_BINARY
		fmt.Print("Enter filename: ")
		filename, _ := reader.ReadString('\n')
		filename = strings.TrimSpace(filename)

		fmt.Print("Enter file path to upload: ")
		filePath, _ := reader.ReadString('\n')
		filePath = strings.TrimSpace(filePath)

		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			return
		}

		fmt.Print("Enter notes (optional): ")
		notes, _ := reader.ReadString('\n')
		notes = strings.TrimSpace(notes)

		secretData = &pb.BinaryData{
			Filename: filename,
			Content:  content,
			Notes:    &notes,
		}
	default:
		fmt.Println("Invalid secret type")
		return
	}

	req := &pb.CreateSecretRequest{
		Name: name,
		Type: secretType,
	}

	switch data := secretData.(type) {
	case *pb.PasswordData:
		req.Data = &pb.CreateSecretRequest_PasswordData{PasswordData: data}
	case *pb.CardData:
		req.Data = &pb.CreateSecretRequest_CardData{CardData: data}
	case *pb.BinaryData:
		req.Data = &pb.CreateSecretRequest_BinaryData{BinaryData: data}
	}

	ctx := withToken(context.Background())
	res, err := secretClient.CreateSecret(ctx, req)
	if err != nil {
		fmt.Printf("Failed to create secret: %v\n", err)
		return
	}

	fmt.Printf("Secret created successfully with ID: %d\n", res.GetId())
}

func getSecrets() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter secret name to filter (leave empty for all): ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	req := &pb.GetSecretRequest{}
	if name != "" {
		req.Name = &name
	}

	ctx := withToken(context.Background())
	res, err := secretClient.GetSecret(ctx, req)
	if err != nil {
		fmt.Printf("Failed to get secrets: %v\n", err)
		return
	}

	if len(res.GetSecrets()) == 0 {
		fmt.Println("No secrets found")
		return
	}

	fmt.Println("\nSecrets:")
	for i, secret := range res.GetSecrets() {
		fmt.Printf("\n%d. Name: %s, Type: %s\n", i+1, secret.GetName(), secret.GetType().String())

		switch secret.GetData().(type) {
		case *pb.GetSecret_PasswordData:
			data := secret.GetPasswordData()
			fmt.Printf("   Username: %s\n", data.GetUsername())
			fmt.Printf("   Password: %s\n", data.GetPassword())
			fmt.Printf("   URL: %s\n", data.GetUrl())
			if data.GetNotes() != "" {
				fmt.Printf("   Notes: %s\n", data.GetNotes())
			}
		case *pb.GetSecret_CardData:
			data := secret.GetCardData()
			fmt.Printf("   Owner: %s\n", data.GetOwner())
			fmt.Printf("   Number: %s\n", maskCardNumber(data.GetNumber()))
			fmt.Printf("   CVV: %s\n", data.GetCVV())
			fmt.Printf("   Expire Date: %s\n", data.GetExpireDate())
			if data.GetNotes() != "" {
				fmt.Printf("   Notes: %s\n", data.GetNotes())
			}
		case *pb.GetSecret_BinaryData:
			data := secret.GetBinaryData()
			fmt.Printf("   Filename: %s\n\n", data.GetFilename())
			fmt.Printf("-------\n\n")
			fmt.Printf("   Content:\n %s\n\n", string(data.GetContent()))
			fmt.Printf("-------\n")

			fmt.Printf("   Size: %d bytes\n", len(data.GetContent()))
			if data.GetNotes() != "" {
				fmt.Printf("   Notes: %s\n", data.GetNotes())
			}
		}
	}
}

func withToken(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", token)
}

func maskCardNumber(number string) string {
	if len(number) <= cardNumberSafeLen {
		return number
	}
	return "**** **** **** " + number[len(number)-4:]
}
