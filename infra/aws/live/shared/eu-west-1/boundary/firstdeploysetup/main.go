package main

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/boundary/api"
	"github.com/hashicorp/boundary/api/accounts"
	"github.com/hashicorp/boundary/api/authmethods"
	"github.com/hashicorp/boundary/api/groups"
	"github.com/hashicorp/boundary/api/roles"
	"github.com/hashicorp/boundary/api/scopes"
	"github.com/hashicorp/boundary/api/users"
	"github.com/hashicorp/boundary/sdk/wrapper"
)


func main () {
	orgName := "infra"
	projectName := "euwest-1"
	orgAdminUsername := "admin"
	orgAdminPassword := os.Getenv("ADMIN_PASSWORD")
	kmsKeyId := os.Getenv("KMS_KEY_ID")

	client, ctx, err := createApiClientFromRecoveryKey(kmsKeyId)
	if err != nil {
		fmt.Println("Failed to create client from recovery key.")
		return
	}
	client.SetAddr("https://boundary.fynbos.dev")

	org, err := createOrg(client, ctx, orgName, "AWS infrastructure.")
	if err != nil {
		fmt.Println("Failed to create Fynbos org.")
		fmt.Println(err)
		return
	}
	fmt.Println(org)

	// TODO: change this to managed groups and connect to Google
	passwordAuth, err := createAuthMethod(client, ctx, org.Id, "password", orgName + "-password-auth", "Password auth for Fynbos")
	if err != nil {
		fmt.Println("Failed to create password auth for organisation.")
		fmt.Println(err)
		return
	}
	fmt.Println(passwordAuth)

	orgAdminRole, err := createOrgAdminRole(client, ctx, org.Id, orgName + "-admin")
	if err != nil {
		fmt.Println("Failed to create org admin role.")
		fmt.Println(err)
		return
	}
	fmt.Println(orgAdminRole)

	// create infrastructure project and give created admin user priviledge 
	project, err := createProject(client, ctx, org.Id, projectName, "eu-west region.")
	if err != nil {
		fmt.Println("Failed to create infrastructure project.")
		fmt.Println(err)
		return
	}
	fmt.Println(project)

	projectAdminRole, err := createProjectAdminRole(client, ctx, org.Id, project.Id, projectName + "-admin")
	if err != nil {
		fmt.Println("Failed to create project admin role.")
		fmt.Println(err)
		return
	}
	fmt.Println(projectAdminRole)

	// create admin group and assign to admin role
	usr, acc, err := createUserAndAccount(client, ctx, org.Id, passwordAuth.Id, orgAdminUsername, orgAdminPassword)
	if err != nil {
		fmt.Println("Failed to create admin user.")
		fmt.Println(err)
		return
	}
	fmt.Println(usr)
	fmt.Println(acc)

	adminGroup, err := createGroup(client, ctx, org.Id, orgName + "-admins", "Read/write access.")
	if err != nil {
		fmt.Println("Failed to create admin group.")
		fmt.Println(err)
		return
	}
	fmt.Println(adminGroup)

	err = addUserToGroup(client, ctx, adminGroup, usr.Id)
	if err != nil {
		fmt.Println("Failed to add admin user to admin group.")
		fmt.Println(err)
		return
	}

	// TODO: for some reason the add principals part fails with the go sdk
	// err = assignRole(client,ctx, projectAdminRole, adminGroup.Id)
	// if err != nil {
	// 	fmt.Println("Failed to assign project admin role to admin group.")
	// 	fmt.Println(err)
	// 	return
	// }

	// TODO: for some reason the add principals part fails with the go sdk. That's why the roles are returned from this function so that the IDs can be printed and used with `assign_roles.sh` script
	globalAnonRole, orgAnonRole, err := allowAnonToListScopesAndAuthMethods(client, ctx, org.Id)
	if err != nil {
		fmt.Println("Failed to allow anonymous user to list orgs and authentication methods.")
		fmt.Println(err)
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

	// TODO: add hosts etc
}

func createApiClientFromRecoveryKey (kmsKeyId string) (*api.Client, context.Context, error) {
	kmsConfig := fmt.Sprintf(`
		kms "awskms" {
		  purpose = "recovery"
		  key_id = "global_recovery"
		  kms_key_id = "%s"
		}
	`, kmsKeyId)
	fmt.Println("creating client")
	client, err := api.NewClient(nil)
	if err != nil { return nil, nil, err }

	fmt.Println("creating wrapper")
	w, err := wrapper.GetWrapperFromHcl(kmsConfig, "recovery")
	if err != nil {	return nil, nil, err }

	fmt.Println("setting recovery key")
	client.SetRecoveryKmsWrapper(w)
	ctx := context.Background()

	return client, ctx, nil
}

func createOrg(client *api.Client, ctx context.Context, name string, description string) (*scopes.Scope ,error) {
	scopesClient := scopes.NewClient(client)

	// Boundary has a default `global` scope which encapsulates all orgs that are created.
	result, err := scopesClient.Create(
		ctx,
		"global",
		scopes.WithName(name),
		scopes.WithDescription(description),
		scopes.WithSkipAdminRoleCreation(true),
		scopes.WithSkipDefaultRoleCreation(true),
	)
	if err != nil { return nil, err }

	return result.Item, err
}

func createProject(client *api.Client, ctx context.Context, orgId string, name string, description string) (*scopes.Scope ,error) {
	scopesClient := scopes.NewClient(client)

	result, err := scopesClient.Create(
		ctx,
		orgId,
		scopes.WithName(name),
		scopes.WithDescription(description),
		scopes.WithSkipAdminRoleCreation(true),
		scopes.WithSkipDefaultRoleCreation(true),
	)
	if err != nil { return nil, err }

	return result.Item, err
}

func listScopes (client *api.Client, ctx context.Context, name string) ([]*scopes.Scope, error) {
	scopesClient := scopes.NewClient(client)
	result, err := scopesClient.List(ctx, name)
	if err != nil { return nil, err }

	return result.Items, nil
}

func createAuthMethod(client *api.Client, ctx context.Context, orgId string, resourceType string, name string, description string) (*authmethods.AuthMethod ,error) {
	authClient := authmethods.NewClient(client)
	result, err := authClient.Create(ctx, "password", orgId, authmethods.WithName(name), authmethods.WithDescription(description))
	if err != nil { return nil, err }

	return result.Item, err
}

func deleteAuthMethod(client *api.Client, ctx context.Context, orgId string) error {
	authClient := authmethods.NewClient(client)
	_, err := authClient.Delete(ctx, orgId)
	if err != nil { return err }

	return nil
}

func listAuthMethods(client *api.Client, ctx context.Context, orgId string) ([]*authmethods.AuthMethod, error) {
	authClient := authmethods.NewClient(client)
	result, err := authClient.List(ctx, orgId)
	if err != nil { return nil, err }

	return result.Items, err
}

func createUserAndAccount(client *api.Client, ctx context.Context, orgId string, authMethodId string, name string, password string) (*users.User, *accounts.Account, error) {
	accountClient := accounts.NewClient(client)
	acc, err := accountClient.Create(ctx, authMethodId, accounts.WithName(name), accounts.WithPasswordAccountPassword(password), accounts.WithPasswordAccountLoginName(name))
	if err != nil { return nil, nil, err }

	usersClient := users.NewClient(client)
	user, err := usersClient.Create(ctx, orgId, users.WithName(name))
	if err != nil { return nil, nil, err }

	_, err = usersClient.SetAccounts(ctx, user.Item.Id, user.Item.Version, []string{acc.Item.Id})
	if err != nil { return nil, nil, err }

	return user.Item, acc.Item, nil
}

func createGroup(client * api.Client, ctx context.Context, orgId string, name string, description string) (*groups.Group ,error) {
	groupsClient := groups.NewClient(client)
	group, err := groupsClient.Create(ctx, orgId, groups.WithDescription(description), groups.WithName(name))
	if err != nil { return nil, err }

	return group.Item, nil
}

func addUserToGroup(client *api.Client, ctx context.Context, group *groups.Group, userId string) error {
	groupsClient := groups.NewClient(client)
	_, err := groupsClient.AddMembers(ctx, group.Id, group.Version, []string{userId})
	if err != nil { return err }

	return nil
}

// This create the orgAdminRole on the `global` scope and scopes the grants to the organisation.
// https://www.boundaryproject.io/docs/installing/no-gen-resources#org-admin-orgAdminRole-for-myuser
func createOrgAdminRole (client *api.Client, ctx context.Context, orgId string, name string) (*roles.Role, error) {
	rolesClient := roles.NewClient(client)
	role, err := rolesClient.Create(ctx, "global", roles.WithName(name), roles.WithGrantScopeId(orgId))
	if err != nil { return nil, err }

	_, err = rolesClient.AddGrants(ctx, role.Item.Id, role.Item.Version, []string{
		"id=*;type=*;actions=*",
	})
	if err != nil { return nil, err	}

	return role.Item, nil
}

func createProjectAdminRole (client *api.Client, ctx context.Context, orgId string, projectId string, name string) (*roles.Role, error) {
	rolesClient := roles.NewClient(client)
	role, err := rolesClient.Create(ctx, orgId, roles.WithName(name), roles.WithGrantScopeId(projectId))
	if err != nil { return nil, err }

	_, err = rolesClient.AddGrants(ctx, role.Item.Id, role.Item.Version, []string{
		"id=*;type=*;actions=*",
	})
	if err != nil { return nil, err	}

	return role.Item, nil
}

func assignRole(client *api.Client, ctx context.Context, role *roles.Role, principalId string) error {
	rolesClient := roles.NewClient(client)
	_, err := rolesClient.AddPrincipals(ctx, role.Id, role.Version, []string{principalId})
	if err != nil { return err	}

	return nil
}

func allowAnonToListScopesAndAuthMethods(client *api.Client, ctx context.Context, orgId string) (*roles.Role, *roles.Role, error) {
	rolesClient := roles.NewClient(client)
	globalAnonListRole, err := rolesClient.Create(ctx, "global", roles.WithName("global_anon_listing"))
	if err != nil { return nil, nil, err }

	_, err = rolesClient.AddGrants(ctx, globalAnonListRole.Item.Id, globalAnonListRole.Item.Version, []string{
		"id=*;type=auth-method;actions=list,authenticate",
		"id=*;type=scope;actions=list,no-op",
		"id={{account.id}};actions=read,change-password",
	})

	// TODO: uncomment once issue with go sdk assigning roles is resolved
	// _, err = rolesClient.AddPrincipals(ctx, globalAnonListRole.Item.Id, globalAnonListRole.Item.Version, []string{"u_anon"})
	// if err != nil { return nil, nil, err	}

	orgAnonListRole, err := rolesClient.Create(ctx, orgId, roles.WithName("org_anon_listing"))
	if err != nil { return nil, nil, err }

	_, err = rolesClient.AddGrants(ctx, orgAnonListRole.Item.Id, orgAnonListRole.Item.Version, []string{
		"id=*;type=auth-method;actions=list,authenticate",
		"id=*;type=scope;actions=list,no-op",
		"id={{account.id}};actions=read,change-password",
	})

	// TODO: uncomment once issue with go sdk assigning roles is resolved
	// _, err = rolesClient.AddPrincipals(ctx, orgAnonListRole.Item.Id, orgAnonListRole.Item.Version, []string{"u_anon"})
	// if err != nil { return nil, nil, err	}

	return globalAnonListRole.Item, orgAnonListRole.Item, nil
}
