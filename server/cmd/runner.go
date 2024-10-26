package cmd

import (
	"github.com/elliotchance/orderedmap/v2"
	"github.com/jamius19/postbranch/logger"
	"os/exec"
)

type Command struct {
	Name string
	Args []string
}

type CommandOutput struct {
	Name   string
	Output *string
	Error  error
}

var log = logger.Logger

func Single(name string, args ...string) (*string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		return nil, err
	}

	output := string(out)

	return &output, nil
}

func Multi(cmds *orderedmap.OrderedMap[string, Command]) (*orderedmap.OrderedMap[string, CommandOutput], error) {
	Log(cmds)

	outputs := orderedmap.NewOrderedMap[string, CommandOutput]()

	for el := cmds.Front(); el != nil; el = el.Next() {
		command := el.Value
		output, err := Single(command.Name, command.Args...)

		if err != nil {
			log.Errorf("Error executing key: %s command: %s args: %s error: %s", el.Key, command.Name, command.Args, err)

			outputs.Set(el.Key, CommandOutput{
				Name:   command.Name,
				Output: nil,
				Error:  err,
			})

			return outputs, err

		}

		outputs.Set(el.Key, CommandOutput{
			Name:   command.Name,
			Output: output,
			Error:  nil,
		})

	}

	return outputs, nil
}

func Get(name string, args ...string) Command {
	return Command{
		Name: name,
		Args: args,
	}
}

func Log(cmds *orderedmap.OrderedMap[string, Command]) {
	log.Infof("Listing commands")

	for el := cmds.Front(); el != nil; el = el.Next() {
		log.Infof("Executing %s command: %s %s", el.Key, el.Value.Name, el.Value.Args)
	}

	log.Infof("Command logging finished")
}