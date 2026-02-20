package tui

import (
	"testing"

	"local-dev-tool/internal/rafiki"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestNewAssetSelectorModel(t *testing.T) {
	assets := []rafiki.Asset{
		{Code: "USD", Scale: 2},
		{Code: "EUR", Scale: 2},
		{Code: "GBP", Scale: 2},
	}

	model := NewAssetSelectorModel(assets)

	assert.Equal(t, 3, len(model.assets))
	assert.Equal(t, 0, model.cursor)
	assert.False(t, model.Cancelled)

	// All should be selected by default
	for i := range assets {
		assert.True(t, model.selected[i])
	}
}

func TestAssetSelectorModel_Navigation(t *testing.T) {
	assets := []rafiki.Asset{
		{Code: "USD", Scale: 2},
		{Code: "EUR", Scale: 2},
		{Code: "GBP", Scale: 2},
	}

	model := NewAssetSelectorModel(assets)

	// Test down navigation
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updatedModel.(AssetSelectorModel)
	assert.Equal(t, 1, model.cursor)

	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updatedModel.(AssetSelectorModel)
	assert.Equal(t, 2, model.cursor)

	// Test bounds (shouldn't go past end)
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updatedModel.(AssetSelectorModel)
	assert.Equal(t, 2, model.cursor)

	// Test up navigation
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = updatedModel.(AssetSelectorModel)
	assert.Equal(t, 1, model.cursor)

	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = updatedModel.(AssetSelectorModel)
	assert.Equal(t, 0, model.cursor)

	// Test bounds (shouldn't go past start)
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyUp})
	model = updatedModel.(AssetSelectorModel)
	assert.Equal(t, 0, model.cursor)
}

func TestAssetSelectorModel_Toggle(t *testing.T) {
	assets := []rafiki.Asset{
		{Code: "USD", Scale: 2},
		{Code: "EUR", Scale: 2},
	}

	model := NewAssetSelectorModel(assets)

	// Initially selected
	assert.True(t, model.selected[0])

	// Toggle off
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeySpace})
	model = updatedModel.(AssetSelectorModel)
	assert.False(t, model.selected[0])

	// Toggle back on
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeySpace})
	model = updatedModel.(AssetSelectorModel)
	assert.True(t, model.selected[0])
}

func TestAssetSelectorModel_SelectAll(t *testing.T) {
	assets := []rafiki.Asset{
		{Code: "USD", Scale: 2},
		{Code: "EUR", Scale: 2},
		{Code: "GBP", Scale: 2},
	}

	model := NewAssetSelectorModel(assets)

	// Deselect all first
	model.selected[0] = false
	model.selected[1] = false
	model.selected[2] = false

	// Select all
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	model = updatedModel.(AssetSelectorModel)

	for i := range assets {
		assert.True(t, model.selected[i])
	}
}

func TestAssetSelectorModel_SelectNone(t *testing.T) {
	assets := []rafiki.Asset{
		{Code: "USD", Scale: 2},
		{Code: "EUR", Scale: 2},
	}

	model := NewAssetSelectorModel(assets)

	// All selected by default
	assert.True(t, model.selected[0])
	assert.True(t, model.selected[1])

	// Select none
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	model = updatedModel.(AssetSelectorModel)

	for i := range assets {
		assert.False(t, model.selected[i])
	}
}

func TestAssetSelectorModel_Quit(t *testing.T) {
	assets := []rafiki.Asset{
		{Code: "USD", Scale: 2},
	}

	model := NewAssetSelectorModel(assets)
	assert.False(t, model.Cancelled)

	updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	model = updatedModel.(AssetSelectorModel)

	assert.True(t, model.Cancelled)
	assert.NotNil(t, cmd)
}

func TestAssetSelectorModel_Confirm(t *testing.T) {
	assets := []rafiki.Asset{
		{Code: "USD", Scale: 2},
	}

	model := NewAssetSelectorModel(assets)

	updatedModel, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	model = updatedModel.(AssetSelectorModel)

	assert.False(t, model.Cancelled)
	assert.NotNil(t, cmd)
}

func TestAssetSelectorModel_SelectedAssets(t *testing.T) {
	assets := []rafiki.Asset{
		{Code: "USD", Scale: 2},
		{Code: "EUR", Scale: 2},
		{Code: "GBP", Scale: 2},
	}

	model := NewAssetSelectorModel(assets)

	// Deselect middle one
	model.selected[1] = false

	selected := model.SelectedAssets()
	assert.Len(t, selected, 2)
	assert.Equal(t, "USD", selected[0].Code)
	assert.Equal(t, "GBP", selected[1].Code)
}

func TestAssetSelectorModel_View(t *testing.T) {
	assets := []rafiki.Asset{
		{Code: "USD", Scale: 2},
		{Code: "EUR", Scale: 2},
	}

	model := NewAssetSelectorModel(assets)

	view := model.View()

	assert.Contains(t, view, "Select assets")
	assert.Contains(t, view, "USD")
	assert.Contains(t, view, "EUR")
	assert.Contains(t, view, "2/2 assets") // All selected
}
