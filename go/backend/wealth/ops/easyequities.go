package ops

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"gitlab.com/fynbos/log"

	"github.com/playwright-community/playwright-go"
	"github.com/xuri/excelize/v2"
)

func initialSetup() {
	err := playwright.Install()
	if err != nil {
		log.Error("failed to install playwright", zap.Error(err))
	}
}

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

func Login(username, password string) (*EasyEquitiesSession, error) {
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

	if strings.EqualFold(page.URL(), "https://identity.openeasy.io/Mfa/Authenticate") {
		resp.credentialsValid = true
		resp.hasMFA = true

		return &resp, nil
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

func DownloadTFSATransactions(userID int64, session *EasyEquitiesSession) (string, error) {
	if session == nil {
		return "", nil
	}

	_, err := session.page.Goto("https://platform.easyequities.io/TransactionHistory")
	if err != nil {
		return "", err
	}

	err = session.page.Locator("//div[@class=\"inactive-tab-account-type\" and text()=\"TFSA\"]").Click()
	if err != nil {
		return "", err
	}

	download, err := session.page.ExpectDownload(func() error {
		return session.page.Locator("[data-role='export']").Click()
	})
	if err != nil {
		return "", err
	}

	fn := fmt.Sprintf("/tmp/wealt_user_%d_%s.xlsx", userID, time.Now().Format("2006_01_02"))
	err = download.SaveAs(fn)
	if err != nil {
		return "", err
	}
	return fn, nil
}

func ParseTXHistory(userID int64, filePath string) ([]EasyEquitiesDeposit, error) {
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

		for _, r := range rows {
			date, _ := time.Parse("2006/01/02", r[0])
			desc := r[1]
			amt, _ := strconv.ParseFloat(r[2], 64)

			if strings.HasPrefix(desc, "EE-") {
				h.Reset()
				h.Write([]byte(fmt.Sprintf("%d_%s_%s_%s", userID, r[0], r[1], r[2])))
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
