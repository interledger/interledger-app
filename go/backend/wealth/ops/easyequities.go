package ops

import (
	"context"
	"fmt"
	"strings"

	"github.com/playwright-community/playwright-go"
)

func setupPlaywright() (playwright.Browser, error) {
	err := playwright.Install()
	if err != nil {
		return nil, err
	}
	pw, err := playwright.Run()
	if err != nil {
		return nil, err
	}
	browser, err := pw.Chromium.Launch()
	if err != nil {
		return nil, err
	}

	return browser, nil
}

type EasyEquitiesSession struct {
	page             playwright.Page
	hasMFA           bool
	credentialsValid bool
}

func Login(ctx context.Context, username, password string) (*EasyEquitiesSession, error) {
	var resp EasyEquitiesSession
	browser, err := setupPlaywright()
	if err != nil {
		return nil, err
	}

	page, err := browser.NewPage()
	if err != nil {
		return nil, err
	}
	resp.page = page

	_, err = page.Goto("https://platform.easyequities.io/Account/SignIn")
	if err != nil {
		return nil, err
	}

	err = page.Locator("#user-identifier-input").Fill(username)
	if err != nil {
		return nil, err
	}

	err = page.GetByLabel("Password").Fill(password)
	if err != nil {
		return nil, err
	}

	err = page.GetByRole("Button", playwright.PageGetByRoleOptions{Name: "Login"}).Click()
	if err != nil {
		return nil, err
	}

	// TODO: delete MFA, only for testing
	if strings.EqualFold(page.URL(), "https://identity.openeasy.io/Mfa/Authenticate") {
		resp.credentialsValid = true
		resp.hasMFA = true
		var mfa string
		fmt.Println("MFA code:")
		fmt.Scanln(&mfa)

		err = page.Locator("input[name='Code']").Fill(mfa)
		if err != nil {
			return nil, err
		}

		err = page.Locator("input[name='RememberMe']").SetChecked(true)
		if err != nil {
			return nil, err
		}
	}

	txt, err := page.Locator(".validation-summary-errors li, #trust-account-types").TextContent()
	if err != nil {
		return nil, err
	}

	fmt.Println(txt, page.URL())
	if strings.Contains(txt, "Credentials supplied are invalid") {
		return &resp, nil
	}

	resp.credentialsValid = true

	return &resp, nil
}

func GetTFSATransactions(ctx context.Context, session *EasyEquitiesSession) ([]byte, error) {
	if session == nil {
		return nil, nil
	}

	_, err := session.page.Goto("https://platform.easyequities.io/TransactionHistory")
	if err != nil {
		return nil, err
	}

	err = session.page.Locator("//div[@class=\"inactive-tab-account-type\" and text()=\"TFSA\"]").Click()
	if err != nil {
		return nil, err
	}

	session.page.OnDownload(func(download playwright.Download) {
		fmt.Println("/tmp/" + download.SuggestedFilename())
		err = download.SaveAs("/tmp/" + download.SuggestedFilename())
	})
	err = session.page.Locator("[data-role='export']").Click()
	if err != nil {
		return nil, err
	}
	return nil, err
}
