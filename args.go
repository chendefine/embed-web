package embedweb

import (
	"os"
	"strconv"
	"strings"
)

func getPortFromCmdArgs() *int {
	port := ParseCmdArgsInt("p", "port", nil)
	if port != nil && *port >= 0 && *port < 65536 {
		return port
	}
	return nil
}

func ParseCmdArgsInt(shortFlag, longFlag string, defaultValue *int) *int {
	args := os.Args
	shortFlag, longFlag = "-"+shortFlag, "--"+longFlag
	for i, arg := range args {
		// 处理 -flag value 或 --flag value 格式
		if arg == shortFlag || arg == longFlag {
			if i+1 < len(args) {
				if val, err := strconv.Atoi(args[i+1]); err == nil {
					return &val
				}
			}
		} else if strings.HasPrefix(arg, longFlag+"=") {
			// 处理 --flag=value 格式
			if val, err := strconv.Atoi(arg[len(longFlag)+1:]); err == nil {
				return &val
			}
		} else if strings.HasPrefix(arg, shortFlag+"=") {
			// 处理 -flag=value 格式
			if val, err := strconv.Atoi(arg[len(shortFlag)+1:]); err == nil {
				return &val
			}
		}
	}
	return defaultValue
}
