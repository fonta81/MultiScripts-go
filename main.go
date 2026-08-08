package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/jroimartin/gocui"
)

// ScriptDef define la estructura escalable para nuestro catálogo de scripts
type ScriptDef struct {
	Name        string
	Description string
	Command     string
	Args        []string
	SourceFile  string
}

// Catálogo global de scripts
var scripts = []ScriptDef{
	{
		Name:        "Test de Red (Ping)",
		Description: "Ejecuta un ping a los servidores de Google (8.8.8.8) para comprobar la conectividad.",
		Command:     "bash",
		Args:        []string{"./test_red.sh"},
		SourceFile:  "./test_red.sh",
	},
	{
		Name:        "Script en Python",
		Description: "Un script de prueba simple escrito en Python que imprime un mensaje.",
		Command:     "python3",
		Args:        []string{"./hola.py"},
		SourceFile:  "./hola.py",
	},
	{
		Name:        "Información del Sistema",
		Description: "Obtiene información básica del kernel y arquitectura del sistema operativo.",
		Command:     "bash",
		Args:        []string{"./sysinfo.sh"},
		SourceFile:  "./sysinfo.sh",
	},
}

var currentIndex = 0

func main() {
	// Generar archivos de prueba automáticamente para que el ejemplo funcione out-of-the-box
	generateTestFiles()

	// Inicializar gocui
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.Highlight = true
	g.Cursor = true
	g.SelFgColor = gocui.ColorGreen

	g.SetManagerFunc(layout)

	if err := keybindings(g); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

// layout define la disposición de los paneles en la terminal
func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	// Panel izquierdo: Lista de scripts (1/3 de la pantalla)
	if v, err := g.SetView("list", 0, 0, maxX/3-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Catálogo de Scripts "
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack

		for _, s := range scripts {
			fmt.Fprintln(v, s.Name)
		}

		if _, err := g.SetCurrentView("list"); err != nil {
			return err
		}
	}

	// Panel superior derecho: Información/Metadatos (1/3 de alto)
	if v, err := g.SetView("info", maxX/3, 0, maxX-1, maxY/3-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Información / Resumen "
		v.Wrap = true
	}

	// Panel inferior derecho: Vista previa de código
	if v, err := g.SetView("preview", maxX/3, maxY/3, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Vista Previa (Código Fuente) "
		v.Wrap = true
	}

	// Forzar la actualización inicial de la vista de detalles
	updateDetails(g)
	return nil
}

// keybindings mapea los atajos de teclado a las funciones
func keybindings(g *gocui.Gui) error {
	// Salir del programa con q o Ctrl+C
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	if err := g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		return err
	}

	// Navegación en la lista (Flechas Arriba/Abajo)
	if err := g.SetKeybinding("list", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("list", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}

	// Ejecutar script con Enter
	if err := g.SetKeybinding("list", gocui.KeyEnter, gocui.ModNone, executeSelected); err != nil {
		return err
	}

	// Cerrar el modal de salida de ejecución
	if err := g.SetKeybinding("modal", gocui.KeyEnter, gocui.ModNone, closeModal); err != nil {
		return err
	}

	return nil
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if currentIndex < len(scripts)-1 {
			if err := v.SetCursor(cx, cy+1); err == nil {
				currentIndex++
				updateDetails(g)
			}
		}
	}
	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if currentIndex > 0 {
			if err := v.SetCursor(cx, cy-1); err == nil {
				currentIndex--
				updateDetails(g)
			}
		}
	}
	return nil
}

// updateDetails actualiza los paneles de Info y Preview según el script seleccionado
func updateDetails(g *gocui.Gui) {
	script := scripts[currentIndex]

	// Actualizar Vista Info
	g.Update(func(gui *gocui.Gui) error {
		vInfo, err := gui.View("info")
		if err != nil {
			return err
		}
		vInfo.Clear()
		fmt.Fprintf(vInfo, "\033[1mNombre:\033[0m %s\n", script.Name)
		fmt.Fprintf(vInfo, "\033[1mComando:\033[0m %s %s\n", script.Command, strings.Join(script.Args, " "))
		fmt.Fprintf(vInfo, "\n\033[1mDescripción:\033[0m\n%s\n", script.Description)
		fmt.Fprintf(vInfo, "\nAtajos: [\u2191/\u2193] Navegar | [Enter] Ejecutar | [q] Salir")
		return nil
	})

	// Actualizar Vista Preview (Primeras 10-15 líneas)
	g.Update(func(gui *gocui.Gui) error {
		vPrev, err := gui.View("preview")
		if err != nil {
			return err
		}
		vPrev.Clear()

		content, err := os.ReadFile(script.SourceFile)
		if err != nil {
			fmt.Fprintf(vPrev, "Error leyendo archivo: %v", err)
			return nil
		}

		lines := strings.Split(string(content), "\n")
		limit := 15
		if len(lines) < limit {
			limit = len(lines)
		}

		for i := 0; i < limit; i++ {
			fmt.Fprintln(vPrev, lines[i])
		}
		if len(lines) > limit {
			fmt.Fprintln(vPrev, "\n... (archivo truncado en la vista previa) ...")
		}
		return nil
	})
}

// executeSelected captura y muestra la salida del script en una ventana modal (Overlay)
func executeSelected(g *gocui.Gui, v *gocui.View) error {
	script := scripts[currentIndex]

	cmd := exec.Command(script.Command, script.Args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	outputStr := out.String()
	if err != nil {
		outputStr += fmt.Sprintf("\n[Error de ejecución]: %v", err)
	}

	// Crear vista modal para mostrar la salida
	maxX, maxY := g.Size()
	if vModal, err := g.SetView("modal", maxX/6, maxY/6, maxX-(maxX/6), maxY-(maxY/6)); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		vModal.Title = fmt.Sprintf(" Salida: %s (Presiona ENTER para cerrar) ", script.Name)
		vModal.Wrap = true
		vModal.Autoscroll = true
		fmt.Fprint(vModal, outputStr)
		g.SetCurrentView("modal")
	}
	return nil
}

func closeModal(g *gocui.Gui, v *gocui.View) error {
	if err := g.DeleteView("modal"); err != nil {
		return err
	}
	if _, err := g.SetCurrentView("list"); err != nil {
		return err
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

// generateTestFiles crea archivos físicos locales para que el código tenga algo que leer y ejecutar
func generateTestFiles() {
	os.WriteFile("test_red.sh", []byte("#!/bin/bash\necho \"Iniciando prueba de red...\"\nping -c 3 8.8.8.8\necho \"Prueba completada.\"\n"), 0o755)
	os.WriteFile("hola.py", []byte("# Script de prueba en Python\nimport time\nprint('Iniciando script...')\ntime.sleep(1)\nprint('¡Hola desde un subcomando gestionado por Go!')\n"), 0o755)
	os.WriteFile("sysinfo.sh", []byte("#!/bin/bash\necho \"--- Información del Sistema ---\"\nuname -smr\necho \"--- Uso de disco ---\"\ndf -h / | tail -n 1\n"), 0o755)
}
