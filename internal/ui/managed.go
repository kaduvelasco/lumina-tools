package ui

import (
	"context"
	"io"
)

// RunManagedSelect runs the standard multiselect→diff→execute flow for managed
// tool domains (IDEs, LLMs, terminals). items must already have Selected set to
// the current install state; installed maps item Label to whether it is installed.
// onInstall and onUninstall receive the original catalogue index.
func RunManagedSelect(
	ctx context.Context,
	stdin io.Reader,
	stdout io.Writer,
	header string,
	items []SelectItem,
	installed map[string]bool,
	onInstall, onUninstall func(idx int) error,
	doneMsg string,
) error {
	finalItems, confirmed, err := RunMultiSelect(ctx, stdin, stdout, items)
	if err != nil {
		return err
	}
	if !confirmed {
		Warning(stdout, "Operação cancelada.")
		WaitEnter(stdout)
		return nil
	}

	type queued struct {
		idx     int
		name    string
		install bool
	}
	var queue []queued
	for i, item := range finalItems {
		name := items[i].Label
		switch {
		case item.Selected && !installed[name]:
			queue = append(queue, queued{i, name, true})
		case !item.Selected && installed[name]:
			queue = append(queue, queued{i, name, false})
		}
	}

	if len(queue) == 0 {
		Info(stdout, "Nenhuma alteração necessária.")
		WaitEnter(stdout)
		return nil
	}

	PrintHeader(stdout, header)

	for _, q := range queue {
		if q.install {
			Info(stdout, "Instalando "+q.name+"...")
			if err := onInstall(q.idx); err != nil {
				Warning(stdout, "Falha ao instalar "+q.name+": "+err.Error())
			} else {
				Success(stdout, q.name+" instalado.")
			}
		} else {
			Info(stdout, "Desinstalando "+q.name+"...")
			if err := onUninstall(q.idx); err != nil {
				Warning(stdout, "Falha ao remover "+q.name+": "+err.Error())
			} else {
				Success(stdout, q.name+" removido.")
			}
		}
	}

	Success(stdout, doneMsg)
	WaitEnter(stdout)
	return nil
}
