package main

import (
	"context"
	"log"

	"github.com/lestrrat-go/jwx/v3/jwk"

	fiantv1 "gitlab.com/fynbos/sdks/fiant/v1"
)

func main() {
	ptiPrivateKey, err := jwk.ParseKey([]byte(`{"d":"UMjijq3vSlU1DjJsvE1CZ8LEQMiHecbj6qv2iuhUNSQOzv5iPG2UDLP3FF2xonmV-ZV9GuQSEN78tSSmuU0Sguse7YWxAVM6w-EeeVS9rq-Uog4IPXaA50Yeu1MM9lwjmXOLDz7RASoh5u54mdg280RoWjDiP-40iB-Ngbv4FpO8CGTa-AqEVxMcXN36GxBLfyyNGAPNwoNyayvNiaAq3FdQ5qi8yhBaw6R43ie7Kk19YfKANMUxbBYM28ZGnlVZuXpY5MGraoYPeCILz6Fni5RWtxFKFCi8O77czwlkzCYMDzZ_j6r7UISWHfIgKXrqf3q_JyufDemFigK92LHMNbNx-OF7bVAze0yQNbdAXg1MQ1uWdaScNM4uKKYkef2j0SUOcMh_hwTlreYayYJPS1kKjxcIOdLzXZVgviWdidFsOT1qEr9ealVA7iYV7iqexhVS8CQm1R5R_Nin-BtczGW-6OInq44oeafw7OxwT33rEqciu5V4PsVw209em4egrmZrv8RrIXALwHL4P3_FlGl2Go_cxCZ-kE4hYuAjajVtr7MMVF2Ku3M8wNcHSeG81s0CMk0LEYNmBwXQZO5pP8tqKbfFv6CWW9wL3FBTxmR9HCGxouCDUuJ-t4Gjys9VzvqdYEs2QzimMAhRw5onzha3gNjwXDr6e8874zS4A-0","dp":"yNTThaMbd3mECn6Mp0S-yU2N6Q58OVsiZzwhNO0bqX-Zk9UtGpHNRT-rbzD7hY3ReAQAPwtRffuUWibqEY4KPaizC3w3B9a1E15lSVjk9WgDMIUjJ_rU7sVrjFjr9ULJmCuKQR42CYDXZNgKil76y_6FYvTPDYG_DDtTMAgWO7pyakBp1_pB3ib-YlddeozpJjKhjCgtNsxTeExEo6W5R_8rYTRbpnloIMJSLlHAefnHH7V9ntHlmCqGN_kK1MRAh2iUtU8Uldv9LdteAAIZGNT755dnuJNZ-1l8pj2WFN4pXSANfWwIfapIfb7aUOAt8C4jjyy7VDGSe8bDrGCS2Q","dq":"dhKzGVziOAuuFS2RIBi5UsoiIGzylFdJjKrfjdmZ0pcFiVmWHaTd8UU7O3BE8hiUXdi3E_G-Q2jiVHZfdLyIAkG2W5pGs_AslKv_skF7HnOXPIInbT5EHbHg_pcfR3y6NeWjjZR71p1KAsOtbUzeqdbLWv3QFYQzbDSHEDmV-fbEYLT3pn0FxojtkKwzVvO3rB35tLgEzjXz5r-Yifapj_ub61mlhpeOka7vFYj5ErkOmiwZema1qDH1PnHel4DN37OsDM38B0YZS0p6OKNXHUpY7HrzFRK8bOd1zg039VZuQoGRbYWpYgiGEFjcONgjsoN1_x4TS2L6zRAv_gJK7w","e":"AQAB","kid":"81c27002-f83b-45ea-9b3f-38096d7ac1ee","kty":"RSA","n":"0kDVpwyOIQFTWSTSbnKaUOAbsE2shES99v3JMQ_MvFMzhhjB7A04npiOAfCBpMJDmRxSBr36nxL588s2-B505KN0KFi3osGW5Af0JosBWKdU-ik85tiB3P9RpAMiWKRPEe9aRoNKlzTI263kNVIkc5CnXYNNsJB9UHfG1lfLkTmjtfT-VOeNvvMndJ9G0cw9CJyGSJOsvY32fZPbal407nHGS5f4yj3XJsG3Nwlq1sgsUdvPJznb4bLjJ_nDBzUpIali1rhRiVUfhyhEveut3ZNNzCGPjdbzu-qn_Q3GFmpSgMC-DWJuTOCVBkE6i1TMNMvsimnGAZJMjFb7cyvlxc6T5-m7ug_Tl_A_hfctbPemZDAGoT97DxQ-HakS54I0qaPDBqalz9tJ9iz-QKEzLencIUaKlRD6CiyC-Izffi5RiLgYEKeesUNbguMhHbLgLedWPje-A6cCHiohTajZNpYCDWc7oU2_w9a0z5cqG2Z-2e_b6Rbi8hp3Q1R94Ym55sg1NssyGDE4xxoC5pw7kwNhnQnSRk5Zozoa4_EyofBV3AQ7lIfHUyeAu8yfdcWuwPTetzjHJfsY94gGV2lqdnNAv4fVjPEoHgQRNllaAWQCqovd-UKnl5sRhtEFB9ENVJw-9gwnwoYB9xvWOkdBdVxuVjnrzmytndMku1nh6I0","p":"6G1ZW9-HLnpBqL5WSCg-addGEKf6ZMFSz_uywSd6Sd2MP6Vx5ziwcuvXMyPxayW8EhuPP6ZiKPZx1wCwKn7QjI3F-QRCx85d0tgz0xLjlYOXl3IslFfee0OM2xzqfb41ifETD_gl78rpnvbmyQRF2YaGb2aX8XcGOr57oUvi3hj7cj4dg1rWwTOu9VbHdp3hdDMsbJ0llvs34Qp0x3cicAs6B6E7Mr7XRNpPdZMr53QdjeGlD_cNAs2q_EDk64IFqI5UzKS-eMKoio3Jny84MmgSEUQdguFJ5ig3Z6SzZMubE3IPk3kgysIhRdCyAdHPk1cMzwPBJd6mK4yg4UKYjw","q":"55PFaErNEU5SXWNA6TjVj4tXgNEFeOqKQRPl3YxzaQTZvCkm67pu0CY30ymkSFVtH0YVmCp9oz8iWwk_GUY4i8EHhpb1jzPop7GFbBnvpoHHYl-I-HOKsOHngMNgLRQnhXr_MIi_naNxGmfnL2MRLJtwr_jaR1QHJLRe0ja7vh5-i06Jurt46e7yZfQvvQDDZAes2a8f2XhGRjA23clXVdZ1xplG-64AynuFLcLOh3OYWlym-QG5ZqpKHinC-y0saoDAsuuuXx9wOYTNvkNmz5WmAJDixcfqfAjQ21F2QV5cO6oSNiCVHafpc7byZihZIsCnN-t8g4ugmkoXmRyjIw","qi":"SzgQlV0cv2EWhlKV_kaoq6aty7TeRCHPH8FH7gd4V-Mf4HOoOuS7mtYk8FehKKwliqmy2YadIZERoHJBkUXiFoAlq_AkuQnFxrTT4Q5Fgm1uPkGCnQf92wXFEGFhzmlogOO4GcEQFfEJw8pN0i4UM42W91gfN2S6AjRP8QGOTjdlcVQbY70StbwG0HxewhvmSfwriqBs-3eoYR46jG6lK7vvascGjjZSqV0MZGGHcA_8hp9u7J1ZPksHAxvQ_nY76e7ZBC3ccGv9GkvkdfLxoviEsaovyWrajSQNCcxrk3MG64pkE9Fz-3ZzVH0ySnKmvLHzT4-OUPlQlwXeSsK3Bg"}`))
	if err != nil {
		log.Fatalln(err)
	}
	clientID := "4ca9befe-05a4-4ee1-bb9f-ed2c4283f56d" // example client ID, replace with your own
	url := "https://api.staging.fiant.io/v1/"

	client, err := fiantv1.NewClient(
		fiantv1.WithBaseURL(url),
		fiantv1.WithClientID(clientID),
		fiantv1.WithDerivedKeys(ptiPrivateKey),
	)
	if err != nil {
		log.Fatalln(err)
	}

	// list all users
	// userList, err := client.Users.ListAll(context.Background()) // example usage
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// for _, user := range userList.EnumerateUsers() {
	// 	log.Printf("User: %+v\n", user)
	// }

	// ~list all users

	// create a user
	// example usage, replace with your own user data
	// user := fiantdto.NewUser(
	// 	fiantdto.WithNewUserUUID(),
	// 	fiantdto.WithUserType(fiantdto.PERSON),
	// 	fiantdto.WithUserStatus(fiantdto.ACTIVE),
	// 	fiantdto.WithUserName(*fiantdto.NewName(
	// 		fiantdto.WithFirstName("Ionescu"),
	// 		fiantdto.WithLastName("Andrei"),
	// 		fiantdto.WithMiddleName("Mihai"),
	// 	)),
	// 	fiantdto.WithUserSourceOfFunds("youtube ad revenue"),
	// 	fiantdto.WithUserAddresses([]fiantdto.Address{
	// 		fiantdto.NewAddress(
	// 			fiantdto.WithAddressStreet("1, main street"),
	// 			fiantdto.WithAddressCity("New Hampshire"),
	// 			fiantdto.WithAddressPostalCode("10005"),
	// 			fiantdto.WithAddressStateCode("US-NH"),
	// 			fiantdto.WithAddressCountry("US"),
	// 			fiantdto.AsDefaultAddress(true),
	// 		),
	// 	}),
	// 	// fiantdto.WithUserPTIMetaData(map[string]any{"key": "value"}),
	// 	// fiantdto.WithUserClientMetaData(map[string]any{"key": "value"}),
	// 	fiantdto.WithUserCountryOfCitizenship("US"),

	// 	fiantdto.WithUserDateOfBirth("1991-01-03"),
	// 	fiantdto.WithUserEmails(
	// 		[]fiantdto.Email{
	// 			fiantdto.NewEmail(
	// 				fiantdto.WithEmailAddress("ionescu.andrei@interledger.com"),
	// 				fiantdto.AsDefaultEmail(true),
	// 			),
	// 		},
	// 	),
	// 	fiantdto.WithUserPhones(
	// 		[]fiantdto.Phone{
	// 			fiantdto.NewPhone(
	// 				"+1 (213) 918-2037",
	// 				fiantdto.WithPhoneType("landline"),
	// 				fiantdto.AsDefaultPhone(true),
	// 			),
	// 		},
	// 	),
	// )

	// createdUser, err := client.Users.Create(context.Background(), user)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// fmt.Printf("Created user: %+v\n", createdUser)

	// ~create a user

	// list a user by ID
	// userID := "f2ef378d-2cf6-414b-9814-a371955014de"
	// user, err := client.Users.Get(context.Background(), userID)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// log.Printf("User with ID %s: %+v\n", userID, user)
	// ~list a user by ID

	// get a token
	// token, err := client.AuthService.GetToken(context.Background(), "https://api.staging.fiant.io/v1/user/f2ef378d-2cf6-414b-9814-a371955014de/wallets", "GET")
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// log.Printf("Token: %+v\n", token)
	// ~get a token

	// get a user assessment

	// userID := "01cc2555-3e86-44eb-83d1-410bc243a6d1"
	// user, err := client.Users.Get(context.Background(), userID)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// assesmentRef, err := client.Users.StartAssessment(context.Background(), user, "ilf_dev_withdrawal")
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// log.Printf("Started assessment: %+v\n", assesmentRef)

	// assessment, err := client.Users.GetAssessment(context.Background(), user)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// log.Printf("User assessment: %+v\n", assessment)

	// ~get a user assessment

	// create a wallet for a user

	// userID := "01cc2555-3e86-44eb-83d1-410bc243a6d1"

	// user, err := client.Users.Get(context.Background(), userID)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// wallet := fiantdto.NewWallet(
	// 	fiantdto.WithWalletCurrency("USD"),
	// 	fiantdto.WithWalletLabel("My USD Wallet 3"),
	// 	fiantdto.WithWalletTypeStandard(),
	// )

	// createdWallet, err := client.Users.CreateWallet(context.Background(), user, wallet)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// log.Printf("Created wallet: %+v\n", createdWallet)

	// ~create a wallet for a user

	// list user wallets

	// userID := "01cc2555-3e86-44eb-83d1-410bc243a6d1"

	// user, err := client.Users.Get(context.Background(), userID)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// userWallets, err := client.Users.ListWallets(context.Background(), user)
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	// log.Printf("User wallets: %+v\n", userWallets)

	// ~list user wallets

	// get a wallet by ID

	userID := "01cc2555-3e86-44eb-83d1-410bc243a6d1"
	walletID := "c95698c2-57a8-44f0-adc5-907b53b93efe"

	user, err := client.UsersService.Get(context.Background(), userID)
	if err != nil {
		log.Fatalln(err)
	}

	wallet, err := client.UsersService.GetWallet(context.Background(), user, walletID)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Wallet with ID %s: %+v\n", walletID, wallet)

	// ~get a wallet by ID

}
