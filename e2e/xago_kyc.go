package main

import (
	"fmt"
	"strings"
	"time"
)

// iFillAndSubmitTheMockxagoiframe fills and submits the MockXago Persona KYC iframe.
func (sc *E2EContext) iFillAndSubmitTheMockxagoiframe() error {
	debugPrintln("\n📝 Filling and submitting MockXago Persona KYC iframe...")

	time.Sleep(500 * time.Millisecond)

	// Get the iframe src for diagnostics
	iframeLocator := sc.page.Locator("iframe").First()
	iframeSrc, _ := iframeLocator.GetAttribute("src")
	debugPrintf("   📍 Iframe src: %s\n", iframeSrc)

	// Set up a listener to capture the postMessage
	debugPrintf("   📍 Setting up message listener...\n")
	_, err := sc.page.Evaluate(`() => {
		window.kycCompleted = false;
		window.addEventListener('message', (e) => {
			console.log('Parent received message:', e.data);
			if (e.data?.type === 'OnboardingCompleted') {
				let parsed;
				try { parsed = JSON.parse(e.data.value || '{}'); } catch(ex) { parsed = {}; }
				if (parsed.applicantStatus === 'submitted') {
					window.kycCompleted = true;
					console.log('MockXago KYC completed message received');
				}
			}
		});
	}`)
	if err != nil {
		debugPrintf("   ⚠️  Failed to set up message listener: %v\n", err)
	}

	// Get the frame locator
	frameLocator := sc.page.FrameLocator("iframe").First()

	// Check if iframe exists
	iframeCount, _ := iframeLocator.Count()
	if iframeCount == 0 {
		return fmt.Errorf("no iframe found on page")
	}

	debugPrintf("   📍 Found iframe, searching for form elements\n")

	// Wait for iframe to be loaded and interactive
	time.Sleep(1 * time.Second)

	// Try to find form inputs in the iframe
	inputs := frameLocator.Locator("input")
	inputCount, _ := inputs.Count()
	debugPrintf("   📍 Found %d input fields in iframe\n", inputCount)

	if inputCount == 0 {
		// Take a screenshot for debugging
		_ = sc.iTakeAScreenshot("mockxago-iframe-no-inputs")
		return fmt.Errorf("no input fields found in MockXago iframe")
	}

	// Fill form fields by name attribute
	for i := 0; i < inputCount; i++ {
		input := inputs.Nth(i)
		name, _ := input.GetAttribute("name")
		inputType, _ := input.GetAttribute("type")
		placeholder, _ := input.GetAttribute("placeholder")

		debugPrintf("   📍 Input %d: type=%s, name=%s, placeholder=%s\n", i, inputType, name, placeholder)

		if inputType == "hidden" {
			continue
		}

		var fillErr error
		switch name {
		case "first_name":
			fillErr = input.Fill("Thabo")
		case "last_name":
			fillErr = input.Fill("Mbeki")
		case "dob":
			fillErr = input.Fill("1990-01-15")
		case "address":
			fillErr = input.Fill("42 Nelson Mandela Drive")
		case "city":
			fillErr = input.Fill("Johannesburg")
		case "country":
			fillErr = input.Fill("South Africa")
		}
		if fillErr != nil {
			return fmt.Errorf("failed to fill field %q: %w", name, fillErr)
		}
	}

	// Take screenshot of filled form before submission
	_ = sc.iTakeAScreenshot("mockxago-kyc-form-filled")

	// Look for the submit button and click it
	buttons := frameLocator.Locator("button")
	buttonCount, _ := buttons.Count()
	debugPrintf("   📍 Found %d buttons in iframe\n", buttonCount)

	buttonClicked := false
	for i := 0; i < buttonCount; i++ {
		button := buttons.Nth(i)
		buttonText, _ := button.TextContent()
		buttonText = strings.TrimSpace(buttonText)

		debugPrintf("   📍 Button %d: %s\n", i, buttonText)

		if strings.Contains(strings.ToLower(buttonText), "submit") {
			debugPrintf("   ✓ Clicking submit button: %s\n", buttonText)
			if err := button.Click(); err != nil {
				return fmt.Errorf("failed to click submit button: %w", err)
			}
			buttonClicked = true
			time.Sleep(500 * time.Millisecond)
			break
		}
	}

	if !buttonClicked {
		_ = sc.iTakeAScreenshot("mockxago-no-submit-button")
		return fmt.Errorf("MockXago KYC iframe submit button not found (found %d buttons)", buttonCount)
	}

	debugPrintf("   ✓ MockXago KYC iframe form submitted\n")
	return nil
}
