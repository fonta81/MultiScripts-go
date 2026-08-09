# MultiScripts

MultiScripts es una herramienta de línea de comandos basada en una Interfaz de Usuario de Terminal (TUI) para gestionar y ejecutar scripts de forma interactiva. Desarrollada en Go utilizando la librería `gocui`.

[Read in English](README.md)

## Características

- **Catálogo de Scripts:** Gestión centralizada de scripts ejecutables.
- **Interfaz TUI:** Navegación intuitiva por terminal.
- **Detalles:** Visualización de metadatos (Categoría, Autor, Descripción, Comando).
- **Vista Previa:** Previsualización del código fuente directamente en la interfaz.
- **Ejecución Interactiva:** Lanza scripts y visualiza su salida sin salir de la herramienta.

## Cómo empezar

1.  Asegúrate de tener instalado [Go](https://golang.org/).
2.  Clona este repositorio.
3.  Ejecuta la aplicación:
    ```bash
    go run main.go
    ```

## Añadir nuevos scripts

Para añadir nuevos scripts, simplemente instancia un nuevo elemento `Script` en el slice global `scriptsCatalog` definido en `main.go`:

```go
{
    Name:        "nombre.sh",
    Description: "Descripción breve.",
    Category:    "Categoría",
    Author:      "Autor",
    FilePath:    "scripts/nombre.sh",
    Command:     []string{"bash", "scripts/nombre.sh"},
},
```

La lógica visual se actualizará automáticamente.

## Scripts incluidos

- `hello.sh`: Script de saludo interactivo.
- `sysinfo.sh`: Muestra información básica del sistema.
- `backup.sh`: Simulación de un proceso de backup.

## Controles

- `↑ / ↓`: Navegar por la lista.
- `Enter`: Ejecutar el script seleccionado.
- `q / Ctrl+C`: Salir de la aplicación.
