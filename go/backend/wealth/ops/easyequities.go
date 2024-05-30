package ops

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"
	"github.com/xuri/excelize/v2"
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

		err = page.GetByRole("Button").Click()
		if err != nil {
			return nil, err
		}
	}

	txt, err := page.Locator(".validation-summary-errors li, #trust-account-types").Nth(0).TextContent()
	if err != nil {
		return nil, err
	}

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

	download, err := session.page.ExpectDownload(func() error {
		fmt.Println("GGGGGGGGGGGGGGGGG")
		return session.page.Locator("[data-role='export']").Click()
	})
	if err != nil {
		return nil, err
	}

	fmt.Println("/tmp/" + download.SuggestedFilename())
	err = download.SaveAs("/tmp/" + download.SuggestedFilename())
	if err != nil {
		return nil, err
	}
	return nil, err
}

type EasyEquitiesDeposit struct {
	Hash        string
	Amount      float64
	Date        time.Time
	Description string
}

func ParseTXHistory(ctx context.Context, filePath string) ([]EasyEquitiesDeposit, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	h := sha256.New()

	var resp []EasyEquitiesDeposit
	sheets := f.GetSheetList()
	for _, s := range sheets {
		rows, err := f.GetRows(s)
		if err != nil {
			return nil, err
		}

		// Chronological order, so we can use the index as part of the hash
		slices.Reverse(rows)
		for i, r := range rows {
			date, _ := time.Parse("2006/01/02", r[0])
			desc := r[1]
			amt, _ := strconv.ParseFloat(r[2], 64)

			if strings.HasPrefix(desc, "EE-") {
				h.Reset()
				h.Write([]byte(fmt.Sprintf("%d_%s_%s_%s", i, r[0], r[1], r[2])))
				hashStr := base64.StdEncoding.EncodeToString(h.Sum(nil))

				resp = append(resp, EasyEquitiesDeposit{
					Hash:        hashStr,
					Amount:      amt,
					Date:        date,
					Description: desc,
				})
			}
		}
	}

	return resp, nil
}
