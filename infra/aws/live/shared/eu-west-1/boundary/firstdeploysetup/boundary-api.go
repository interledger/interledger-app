package main

import (
	"context"
	"fmt"

	"github.com/hashicorp/boundary/api"
	"github.com/hashicorp/boundary/api/accounts"
	"github.com/hashicorp/boundary/api/authmethods"
	"github.com/hashicorp/boundary/api/groups"
	"github.com/hashicorp/boundary/api/roles"
	"github.com/hashicorp/boundary/api/scopes"
	"github.com/hashicorp/boundary/api/users"
	"github.com/hashicorp/boundary/sdk/wrapper"
	"github.com/hashicorp/boundary/api/hostcatalogs"
	"github.com/hashicorp/boundary/api/hosts"
	"github.com/hashicorp/boundary/api/hostsets"
	"github.com/hashicorp/boundary/api/targets"
)


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

func createHostCatalog (ctx context.Context, client *api.Client, name string, projectId string) (*hostcatalogs.HostCatalog, error) {
	catalogClient := hostcatalogs.NewClient(client)

	cg, err := catalogClient.Create(
		ctx,
		"static",
		projectId,
		hostcatalogs.WithName(name),
	)
	if err != nil {
		return nil, err
	}

	return cg.Item, nil
}

func createHost(ctx context.Context, client *api.Client, hostCatalogId string, name string, address string) (*hosts.Host, error) {
	hostClient := hosts.NewClient(client)

	host, err := hostClient.Create(ctx, hostCatalogId, hosts.WithName(name), hosts.WithStaticHostAddress(address))
	if err != nil {
		return nil, err
	}

	return host.Item, nil
}

func createHostSet(ctx context.Context, client *api.Client, hostCatalogId string, name string) (*hostsets.HostSet, error) {
	hostsetClient := hostsets.NewClient(client)

	hostset, err := hostsetClient.Create(ctx, hostCatalogId, hostsets.WithName(name))
	if err != nil {
		return nil, err
	}

	return hostset.Item, nil
}

func addHostToHostSet(ctx context.Context, client *api.Client, hostset *hostsets.HostSet, hostId string) (*hostsets.HostSet, error) {
	hostsetClient := hostsets.NewClient(client)

	result, err := hostsetClient.AddHosts(ctx, hostset.Id, hostset.Version, []string{hostId})
	if err != nil {
		return nil, err
	}

	return result.Item, nil
}

func createTarget (ctx context.Context, client *api.Client, projectId string, name string, targetType string, port uint32, connectionLimit int32) (*targets.Target, error) {
	targetClient := targets.NewClient(client)

	target, err := targetClient.Create(ctx, targetType, projectId, targets.WithName(name), targets.WithTcpTargetDefaultPort(port), targets.WithSessionConnectionLimit(connectionLimit))
	if err != nil {
		return nil, err
	}

	return target.Item, nil
}

func addHostSetToTarget(ctx context.Context, client *api.Client, target *targets.Target, hostSetId string) (*targets.Target, error) {
	targetClient := targets.NewClient(client)

	result, err := targetClient.AddHostSets(ctx, target.Id, target.Version, []string{hostSetId})
	if err != nil {
		return nil, err
	}

	return result.Item, nil
}