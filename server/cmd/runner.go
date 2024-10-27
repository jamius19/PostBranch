package cmd

import (
	"github.com/elliotchance/orderedmap/v2"
	"github.com/jamius19/postbranch/logger"
	"github.com/jamius19/postbranch/util"
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

func Single(key string, name string, args ...string) (*string, error) {
	log.Infof("Executing %s command: %s %v", key, name, args)

	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		log.Errorf("Error executing %s command: %s %v error: %s, output: %v", key, name, args, err, out)
		return nil, err
	}

	output := string(out)
	log.Infof("Output for %s: %v", key, output)

	return &output, nil
}

func Multi(cmds *orderedmap.OrderedMap[string, Command]) (*orderedmap.OrderedMap[string, CommandOutput], error) {
	LogCmds(cmds)

	outputs := orderedmap.NewOrderedMap[string, CommandOutput]()

	for el := cmds.Front(); el != nil; el = el.Next() {
		command := el.Value
		output, err := Single(el.Key, command.Name, command.Args...)

		if err != nil {
			log.Errorf("Error executing key: %s command: %s args: %s error: %s, output: %v", el.Key, command.Name, command.Args, err, output)

			outputs.Set(el.Key, CommandOutput{
				Name:   command.Name,
				Output: output,
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

	LogOutputs(outputs)
	return outputs, nil
}

func Get(name string, args ...string) Command {
	return Command{
		Name: name,
		Args: args,
	}
}

func LogCmds(cmds *orderedmap.OrderedMap[string, Command]) {
	log.Infof("--- Logging commands")

	for el := cmds.Front(); el != nil; el = el.Next() {
		log.Infof("Executing %s command: %s %s", el.Key, el.Value.Name, el.Value.Args)
	}

	log.Infof("--- Command logging finished")
}

func LogOutputs(outputs *orderedmap.OrderedMap[string, CommandOutput]) {
	log.Infof("--- Logging outputs")

	for el := outputs.Front(); el != nil; el = el.Next() {
		log.Infof("Output for %s: %v", el.Key, util.StringVal(el.Value.Output))
	}

	log.Infof("--- Output logging finished")
}
