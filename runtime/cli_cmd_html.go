package main

// HtmlCmd implements the Command interface for the legacy --html action.
type HtmlCmd struct{}

func (c *HtmlCmd) Name() string {
	return "--html"
}

func (c *HtmlCmd) Execute(brainRoot string, args []string) error {
	brain := scanBrain(brainRoot)
	result := runSubsumption(brain)
	generateBrainJSON(brainRoot, brain, result)
	return nil
}
