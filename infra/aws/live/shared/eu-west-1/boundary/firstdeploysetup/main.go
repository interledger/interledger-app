package main

import (
	"fmt"
	"os"
	"log"
)

func main () {
	orgName := "infra"
	projectName := "euwest-1"
	orgAdminUsername := "admin"
	orgAdminPassword := os.Getenv("ADMIN_PASSWORD")
	kmsKeyId := os.Getenv("KMS_KEY_ID")

	client, ctx, err := createApiClientFromRecoveryKey(kmsKeyId)
	if err != nil {
		log.Fatalln("Failed to create client from recovery key.")
		return
	}
	client.SetAddr("https://boundary.fynbos.dev")

	org, err := createOrg(client, ctx, orgName, "AWS infrastructure.")
	if err != nil {
		log.Fatalln("Failed to create Fynbos org.")
		log.Fatalln(err)
		return
	}
	log.Println(org)

	// TODO: change this to managed groups and connect to Google
	passwordAuth, err := createAuthMethod(client, ctx, org.Id, "password", orgName + "-password-auth", "Password auth for Fynbos")
	if err != nil {
		log.Fatalln("Failed to create password auth for organisation.")
		log.Fatalln(err)
		return
	}
	log.Println(passwordAuth)

	orgAdminRole, err := createOrgAdminRole(client, ctx, org.Id, orgName + "-admin")
	if err != nil {
		log.Fatalln("Failed to create org admin role.")
		log.Fatalln(err)
		return
	}
	log.Println(orgAdminRole)

	// create infrastructure project and give created admin user priviledge 
	project, err := createProject(client, ctx, org.Id, projectName, "eu-west region.")
	if err != nil {
		log.Fatalln("Failed to create infrastructure project.")
		log.Fatalln(err)
		return
	}
	log.Println(project)

	projectAdminRole, err := createProjectAdminRole(client, ctx, org.Id, project.Id, projectName + "-admin")
	if err != nil {
		log.Fatalln("Failed to create project admin role.")
		log.Fatalln(err)
		return
	}
	log.Fatalln(projectAdminRole)

	// create admin group and assign to admin role
	usr, acc, err := createUserAndAccount(client, ctx, org.Id, passwordAuth.Id, orgAdminUsername, orgAdminPassword)
	if err != nil {
		log.Fatalln("Failed to create admin user.")
		log.Fatalln(err)
		return
	}
	log.Println(usr)
	log.Println(acc)

	adminGroup, err := createGroup(client, ctx, org.Id, orgName + "-admins", "Read/write access.")
	if err != nil {
		fmt.Println("Failed to create admin group.")
		fmt.Println(err)
		return
	}
	log.Println(adminGroup)

	err = addUserToGroup(client, ctx, adminGroup, usr.Id)
	if err != nil {
		log.Fatalln("Failed to add admin user to admin group.")
		log.Fatalln(err)
		return
	}

	// TODO: for some reason the add principals part fails with the go sdk
	err = assignRole(client,ctx, projectAdminRole, adminGroup.Id)
	if err != nil {
		log.Fatalln("Failed to assign project admin role to admin group.")
		log.Fatalln(err)
		return
	}

	// TODO: for some reason the add principals part fails with the go sdk. That's why the roles are returned from this function so that the IDs can be printed and used with `assign_roles.sh` script
	globalAnonRole, orgAnonRole, err := allowAnonToListScopesAndAuthMethods(client, ctx, org.Id)
	if err != nil {
		log.Fatalln("Failed to allow anonymous user to list orgs and authentication methods.")
		log.Fatalln(err)
		return
	}

	fmt.Printf("\nexport ORG_ID=%s; export ADMIN_GROUP_ID=%s; export ORG_ADMIN_ROLE_ID=%s; export PROJECT_ADMIN_ROLE_ID=%s; export GLOBAL_ANON_ROLE_ID=%s; export ORG_ANON_ROLE_ID=%s;\n", 
		org.Id,
		adminGroup.Id,
		orgAdminRole.Id,
		projectAdminRole.Id,
		globalAnonRole.Id,
		orgAnonRole.Id,
	)
	fmt.Println("Now run assign-roles.sh to complete setup.")

	cg, err := createHostCatalog(ctx, client, "private-subnet", project.Id)
	if err != nil {
		log.Fatalln("Failed to create host catalog.")
		log.Fatalln(err)
		return
	}
	log.Println(cg)

	err = createVaultApiTarget(ctx, client, cg)
	if err != nil {
		log.Fatalln(err)
		return
	}
}
