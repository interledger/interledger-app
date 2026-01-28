package main

func allScenarios() []scenario {
	return []scenario{
		scenarioLogin(),
		scenarioCreateAndFetchSubAccount(),
		scenarioUpdateSubAccount(),
		scenarioGetBalance(),
		scenarioBalanceAuthRequired(),
		scenarioListCurrencies(),
	}
}
