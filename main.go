package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/jroimartin/gocui"
)

// ============================================================================
// ESTRUCTURA ESCALABLE: Script
// ============================================================================

// Script representa un script/módulo ejecutable del catálogo.
// Para añadir nuevos scripts, solo instancia más elementos en el slice
// global `scriptsCatalog` — la lógica visual NO necesita cambios.
type Script struct {
	Name        string   // Nombre visible en la lista
	Description string   // Descripción breve
	Category    string   // Categoría (System, DevOps, Utils, etc.)
	Author      string   // Autor del script
	FilePath    string   // Ruta al archivo fuente
	Command     []string // Comando a ejecutar (ej: ["bash", "scripts/hello.sh"])
}

// scriptsCatalog es el catálogo central. Añade nuevos scripts aquí.
var scriptsCatalog = []Script{
	{
		Name:        "hello.sh",
		Description: "Script de saludo interactivo. Muestra un mensaje de bienvenida con la fecha actual y variables de entorno.",
		Category:    "Utils",
		Author:      "DevTeam",
		FilePath:    "scripts/hello.sh",
		Command:     []string{"bash", "scripts/hello.sh"},
	},
	{
		Name:        "sysinfo.sh",
		Description: "Recopila información del sistema: SO, arquitectura, CPU, memoria disponible y uptime.",
		Category:    "System",
		Author:      "DevTeam",
		FilePath:    "scripts/sysinfo.sh",
		Command:     []string{"bash", "scripts/sysinfo.sh"},
	},
	{
		Name:        "backup.sh",
		Description: "Simula un proceso de backup mostrando progreso por pasos. Útil como plantilla para pipelines de respaldo.",
		Category:    "DevOps",
		Author:      "DevTeam",
		FilePath:    "scripts/backup.sh",
		Command:     []string{"bash", "scripts/backup.sh"},
	},
}

// ============================================================================
// ESTADO GLOBAL DE LA APLICACIÓN
// ============================================================================

type AppState struct {
	selectedIndex int      // Índice del script seleccionado en la lista
	scripts       []Script // Referencia al catálogo
	outputBuffer  string   // Buffer para mostrar salida de ejecución
}

var state = &AppState{
	selectedIndex: 0,
	scripts:       scriptsCatalog,
}

// Constantes de nombres de vistas
const (
	ViewList    = "list"
	ViewInfo    = "info"
	ViewPreview = "preview"
	ViewHelp    = "help"
	ViewOutput  = "output"
)

// Colores ANSI para gocui (usando atributos)
var (
	colorTitle    = gocui.ColorYellow | gocui.AttrBold
	colorSelected = gocui.ColorCyan | gocui.AttrBold
	colorBorder   = gocui.ColorWhite
	colorHelp     = gocui.ColorGreen
)

// ============================================================================
// LAYOUT: Posicionamiento de vistas
// ============================================================================

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if maxX < 40 || maxY < 15 {
		return nil // Pantalla muy pequeña
	}

	// Panel izquierdo: Lista de scripts (30% del ancho)
	listWidth := maxX / 3
	if v, err := g.SetView(ViewList, 0, 0, listWidth, maxY-3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Scripts "
		v.Highlight = true
		v.SelBgColor = gocui.ColorCyan
		v.SelFgColor = gocui.ColorBlack
		v.Frame = true
		v.Editable = false
		v.Wrap = false
		updateListView(v)
		if _, err := g.SetCurrentView(ViewList); err != nil {
			return err
		}
	}

	// Panel superior derecho: Info/Resumen
	infoStartX := listWidth + 1
	if v, err := g.SetView(ViewInfo, infoStartX, 0, maxX-1, maxY/3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Info / Resumen "
		v.Frame = true
		v.Wrap = true
		v.Editable = false
		updateInfoView(v)
	}

	// Panel inferior derecho: Vista previa de código
	previewStartY := maxY/3 + 1
	if v, err := g.SetView(ViewPreview, infoStartX, previewStartY, maxX-1, maxY-3); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Vista Previa (10-15 líneas) "
		v.Frame = true
		v.Wrap = false
		v.Editable = false
		updatePreviewView(v)
	}

	// Barra inferior: Atajos de teclado
	if v, err := g.SetView(ViewHelp, 0, maxY-3, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.Editable = false
		v.BgColor = gocui.ColorBlack
		v.FgColor = gocui.ColorGreen
		fmt.Fprintf(v, " ↑/↓: Navegar  |  Enter: Ejecutar  |  q/Ctrl+C: Salir  |  Total: %d scripts ", len(state.scripts))
	}

	return nil
}

// ============================================================================
// ACTUALIZACIÓN DE VISTAS
// ============================================================================

func updateListView(v *gocui.View) {
	v.Clear()
	for i, s := range state.scripts {
		if i == state.selectedIndex {
			fmt.Fprintf(v, "▶ %s\n", s.Name)
		} else {
			fmt.Fprintf(v, "  %s\n", s.Name)
		}
	}
}

func updateInfoView(v *gocui.View) {
	v.Clear()
	if state.selectedIndex < 0 || state.selectedIndex >= len(state.scripts) {
		return
	}
	s := state.scripts[state.selectedIndex]
	fmt.Fprintf(v, "Nombre:      %s\n", s.Name)
	fmt.Fprintf(v, "Categoría:   %s\n", s.Category)
	fmt.Fprintf(v, "Autor:       %s\n", s.Author)
	fmt.Fprintf(v, "Comando:     %s\n", strings.Join(s.Command, " "))
	fmt.Fprintf(v, "Archivo:     %s\n", s.FilePath)
	fmt.Fprintln(v, strings.Repeat("─", 40))
	fmt.Fprintf(v, "Descripción:\n%s\n", s.Description)
}

func updatePreviewView(v *gocui.View) {
	v.Clear()
	if state.selectedIndex < 0 || state.selectedIndex >= len(state.scripts) {
		return
	}
	s := state.scripts[state.selectedIndex]

	lines, err := readFirstLines(s.FilePath, 15)
	if err != nil {
		fmt.Fprintf(v, "[Error leyendo archivo: %v]\n", err)
		return
	}

	for _, line := range lines {
		// Resaltar comentarios en el preview
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") {
			fmt.Fprintf(v, "\x1b[32m%s\x1b[0m\n", line) // Verde para comentarios
		} else {
			fmt.Fprintf(v, "%s\n", line)
		}
	}
	if len(lines) == 15 {
		fmt.Fprintln(v, "\x1b[90m... (truncado)\x1b[0m")
	}
}

// readFirstLines lee las primeras n líneas de un archivo.
func readFirstLines(path string, n int) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) >= n {
			break
		}
	}
	return lines, scanner.Err()
}

// refreshAllViews fuerza la actualización de todas las vistas dinámicas.
func refreshAllViews(g *gocui.Gui) {
	if v, err := g.View(ViewList); err == nil {
		updateListView(v)
	}
	if v, err := g.View(ViewInfo); err == nil {
		updateInfoView(v)
	}
	if v, err := g.View(ViewPreview); err == nil {
		updatePreviewView(v)
	}
}

// ============================================================================
// KEYBINDINGS Y NAVEGACIÓN
// ============================================================================

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if state.selectedIndex < len(state.scripts)-1 {
		state.selectedIndex++
		refreshAllViews(g)
	}
	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	if state.selectedIndex > 0 {
		state.selectedIndex--
		refreshAllViews(g)
	}
	return nil
}

// executeScript lanza el script seleccionado usando os/exec.
func executeScript(g *gocui.Gui, v *gocui.View) error {
	if state.selectedIndex < 0 || state.selectedIndex >= len(state.scripts) {
		return nil
	}
	s := state.scripts[state.selectedIndex]

	// Mostrar vista de salida temporal
	maxX, maxY := g.Size()
	outputView, err := g.SetView(ViewOutput, maxX/6, maxY/6, maxX-maxX/6, maxY-maxY/6)
	if err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		outputView.Title = fmt.Sprintf(" Ejecutando: %s ", s.Name)
		outputView.Frame = true
		outputView.Wrap = true
		outputView.Autoscroll = true
		outputView.BgColor = gocui.ColorBlack
		outputView.FgColor = gocui.ColorWhite
	}

	outputView.Clear()
	fmt.Fprintf(outputView, "▶ Comando: %s\n", strings.Join(s.Command, " "))
	fmt.Fprintf(outputView, "▶ Directorio: %s\n\n", getWorkingDir())

	// Ejecutar el comando
	cmd := exec.Command(s.Command[0], s.Command[1:]...)
	cmd.Dir = getWorkingDir()

	// Capturar stdout y stderr combinados
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(outputView, "\x1b[31m[Error de ejecución: %v]\x1b[0m\n", err)
	}
	fmt.Fprint(outputView, string(output))
	fmt.Fprintln(outputView, "\n\n\x1b[33m[Presiona cualquier tecla para cerrar esta ventana]\x1b[0m")

	// Cambiar foco a la vista de salida
	g.SetCurrentView(ViewOutput)

	// Keybinding temporal para cerrar la vista de salida
	g.SetKeybinding(ViewOutput, gocui.KeyEnter, gocui.ModNone, closeOutputView)
	g.SetKeybinding(ViewOutput, 'q', gocui.ModNone, closeOutputView)
	g.SetKeybinding(ViewOutput, gocui.KeyEsc, gocui.ModNone, closeOutputView)

	return nil
}

func closeOutputView(g *gocui.Gui, v *gocui.View) error {
	g.DeleteView(ViewOutput)
	g.DeleteKeybindings(ViewOutput)
	g.SetCurrentView(ViewList)
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func getWorkingDir() string {
	dir, _ := os.Getwd()
	return dir
}

// ============================================================================
// MAIN
// ============================================================================

func main() {
	// Verificar que los scripts existen
	for _, s := range scriptsCatalog {
		if _, err := os.Stat(s.FilePath); os.IsNotExist(err) {
			log.Printf("Advertencia: No se encontró %s. Créalo antes de ejecutar.\n", s.FilePath)
		}
	}

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)
	g.Cursor = false
	g.Mouse = false
	g.InputEsc = true

	// Keybindings globales
	if err := g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", gocui.KeyEnter, gocui.ModNone, executeScript); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	// Bucle principal
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}
