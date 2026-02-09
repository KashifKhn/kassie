package upgrade

import (
	"fmt"
	"strings"
)

const (
	Cyan   = "\033[0;36m"
	Green  = "\033[0;32m"
	Red    = "\033[0;31m"
	Yellow = "\033[1;33m"
	Muted  = "\033[0;2m"
	Bold   = "\033[1m"
	NC     = "\033[0m"
)

func printLogo() {
	fmt.Println()
	fmt.Printf("%s██╗  ██╗ █████╗ ███████╗███████╗██╗███████╗%s\n", Cyan, NC)
	fmt.Printf("%s██║ ██╔╝██╔══██╗██╔════╝██╔════╝██║██╔════╝%s\n", Cyan, NC)
	fmt.Printf("%s█████╔╝ ███████║███████╗███████╗██║█████╗%s\n", Cyan, NC)
	fmt.Printf("%s██╔═██╗ ██╔══██║╚════██║╚════██║██║██╔══╝%s\n", Cyan, NC)
	fmt.Printf("%s██║  ██╗██║  ██║███████║███████║██║███████╗%s\n", Cyan, NC)
	fmt.Printf("%s╚═╝  ╚═╝╚═╝  ╚═╝╚══════╝╚══════╝╚═╝╚══════╝%s\n", Cyan, NC)
	fmt.Println()
}

func printTagline() {
	fmt.Printf("    %sModern Cassandra & ScyllaDB Explorer%s\n", Muted, NC)
	fmt.Printf("        %sTerminal & Web Interfaces%s\n", Muted, NC)
	fmt.Println()
}

func printHeader(title string) {
	border := strings.Repeat("━", 50)
	fmt.Printf("\n%s━━━ %s %s%s\n\n", Cyan, title, border[:45-len(title)], NC)
}

func printVersionTransition(current, latest string) {
	fmt.Printf("  %sCurrent:%s  %s\n", Muted, NC, current)
	fmt.Printf("  %sLatest:%s   %s\n", Muted, NC, latest)
	fmt.Println()
}

func printProgress(downloaded, total int64, width int) string {
	if total <= 0 {
		return ""
	}

	percent := float64(downloaded) / float64(total) * 100
	if percent > 100 {
		percent = 100
	}

	filled := int(percent / 100 * float64(width))
	empty := width - filled

	bar := strings.Repeat("■", filled) + strings.Repeat("･", empty)
	return fmt.Sprintf("[%s%s%s] %3.0f%%", Cyan, bar, NC, percent)
}

func printStepComplete(message string) {
	fmt.Printf("  %s✓%s %s\n", Green, NC, message)
}

func printAlreadyLatest(version string) {
	fmt.Printf("\n  %s→%s Already on latest version: %s%s%s\n\n", Cyan, NC, Bold, version, NC)
}

func printUpdateAvailable(current, latest string) {
	fmt.Printf("\n  %s!%s Update available: %s → %s%s%s%s\n", Yellow, NC, current, Bold, latest, NC, "\n")
	fmt.Printf("  %s→%s Run %skassie upgrade%s to install\n\n", Cyan, NC, Bold, NC)
}

func printMethod(method string) {
	fmt.Printf("  %sMethod:%s     %s\n", Muted, NC, method)
}

func printDone() {
	border := strings.Repeat("━", 50)
	fmt.Printf("\n%s━━━ Done %s%s\n", Cyan, border, NC)
}

func clearLine() {
	fmt.Print("\r\033[K")
}
