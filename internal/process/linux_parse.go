package process

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/peterintech/nosleepp/internal/agent"
)

const defaultLinuxClockTicksPerSecond = 100

func parseLinuxProcStat(statLine string, clockTicksPerSecond int) (agent.Process, error) {
	if clockTicksPerSecond <= 0 {
		clockTicksPerSecond = defaultLinuxClockTicksPerSecond
	}

	open := strings.Index(statLine, "(")
	close := strings.LastIndex(statLine, ")")
	if open < 0 || close <= open {
		return agent.Process{}, errors.New("invalid proc stat: missing process name")
	}

	pid, err := strconv.Atoi(strings.TrimSpace(statLine[:open]))
	if err != nil {
		return agent.Process{}, err
	}

	name := statLine[open+1 : close]
	fields := strings.Fields(strings.TrimSpace(statLine[close+1:]))
	if len(fields) < 13 {
		return agent.Process{}, errors.New("invalid proc stat: missing CPU fields")
	}

	parentPID, err := strconv.Atoi(fields[1])
	if err != nil {
		return agent.Process{}, err
	}
	userTicks, err := strconv.ParseInt(fields[11], 10, 64)
	if err != nil {
		return agent.Process{}, err
	}
	systemTicks, err := strconv.ParseInt(fields[12], 10, 64)
	if err != nil {
		return agent.Process{}, err
	}

	return agent.Process{
		PID:       pid,
		ParentPID: parentPID,
		Name:      name,
		CPUTime:   ticksToDuration(userTicks+systemTicks, clockTicksPerSecond),
	}, nil
}

func ticksToDuration(ticks int64, clockTicksPerSecond int) time.Duration {
	if ticks <= 0 || clockTicksPerSecond <= 0 {
		return 0
	}
	return time.Duration(ticks) * time.Second / time.Duration(clockTicksPerSecond)
}
