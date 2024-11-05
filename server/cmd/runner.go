package cmd

import (
	"github.com/elliotchance/orderedmap/v2"
	"github.com/jamius19/postbranch/logger"
	"github.com/jamius19/postbranch/util"
	"github.com/jamius19/postbranch/web/responseerror"
	"os/exec"
	"strings"
	"syscall"
)

type Command struct {
	Name      string
	Args      []string
	Sensitive bool
}

type CommandOutput struct {
	Name      string
	Sensitive bool
	Output    *string
	Error     error
}

const EmptyOutput = ""

var log = logger.Logger

func Single(key string, skipLog bool, sensitive bool, name string, args ...string) (*string, error) {
	if !skipLog {
		if !sensitive {
			log.Infof("[s] Executing %s command: %s %v", key, name, args)
		} else {
			log.Infof("[s] Executing %s command: %s *****", key, name)
		}
	}

	cmd := exec.Command(name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	out, err := cmd.CombinedOutput()

	if err != nil {
		if !skipLog {
			if !sensitive {
				log.Errorf("[s] Error executing %s command: %s %v error: %s, output: %s", key, name, args, err, out)
			} else {
				log.Errorf("[s] Error executing %s command: %s ***** error: %s", key, name, err)
			}
		}

		outputStr := string(out)
		return &outputStr, responseerror.Clarify("Error executing command")
	}

	output := string(out)
	if !skipLog {
		if !sensitive {
			log.Infof("[s] Output for %s command: %s, output: %s", key, name, util.SafeStringVal(&output))
		} else {
			log.Infof("[s] Output for %s command: %s, output: *****", key, name)
		}
	}

	return &output, nil
}

func Multi(cmds *orderedmap.OrderedMap[string, Command]) (*orderedmap.OrderedMap[string, CommandOutput], error) {
	LogCmds(cmds)

	outputs := orderedmap.NewOrderedMap[string, CommandOutput]()

	for el := cmds.Front(); el != nil; el = el.Next() {
		command := el.Value
		output, err := Single(el.Key, true, el.Value.Sensitive, command.Name, command.Args...)

		if err != nil {
			log.Errorf(
				"[m] Error executing key: %s command: %s args: %s error: %s, output: %s",
				el.Key, command.Name, command.Args, err, util.SafeStringVal(output),
			)

			outputs.Set(el.Key, CommandOutput{
				Name:      command.Name,
				Output:    output,
				Error:     err,
				Sensitive: command.Sensitive,
			})

			return outputs, err

		}

		outputs.Set(el.Key, CommandOutput{
			Name:      command.Name,
			Output:    output,
			Error:     nil,
			Sensitive: command.Sensitive,
		})

	}

	LogOutputs(outputs)
	return outputs, nil
}

func Get(name string, args ...string) Command {
	return Command{
		Name:      name,
		Args:      args,
		Sensitive: false,
	}
}

func GetSensitive(name string, args ...string) Command {
	return Command{
		Name:      name,
		Args:      args,
		Sensitive: true,
	}
}

func LogCmds(cmds *orderedmap.OrderedMap[string, Command]) {
	log.Infof("[m] >>> Logging commands")

	for el := cmds.Front(); el != nil; el = el.Next() {
		if el.Value.Sensitive {
			log.Infof("[m] Executing %s command: %s *****", el.Key, el.Value.Name)
			continue
		}

		log.Infof("[m] Executing %s command: %s %s", el.Key, el.Value.Name, el.Value.Args)
	}

	log.Infof("[m] <<< Command logging finished")
}

func LogOutputs(outputs *orderedmap.OrderedMap[string, CommandOutput]) {
	log.Infof("[m] >>> Logging outputs")

	for el := outputs.Front(); el != nil; el = el.Next() {
		if el.Value.Sensitive {
			log.Infof("[m] Output for %s: ******", el.Key)
			continue
		}

		log.Infof("[m] Output for %s: %s", el.Key, util.SafeStringVal(el.Value.Output))
	}

	log.Infof("[m] <<< Output logging finished")
}

func GetError(output *orderedmap.OrderedMap[string, CommandOutput]) string {
	errStrBuilder := strings.Builder{}

	if output != nil {
		for el := output.Front(); el != nil; el = el.Next() {
			errStrBuilder.WriteString(el.Key)
			errStrBuilder.WriteString("> ")
			errStrBuilder.WriteString(el.Value.Name)
			errStrBuilder.WriteString(": ")
			errStrBuilder.WriteString(util.SafeStringVal(el.Value.Output))
			errStrBuilder.WriteString(";  ")
		}
	} else {
		errStrBuilder.WriteString("<nil>")
	}

	return errStrBuilder.String()
}
