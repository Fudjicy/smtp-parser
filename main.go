package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

const (
	ColorReset     = "\033[0m"
	ColorRed       = "\033[31m"
	ColorGreen     = "\033[32m"
	ColorYellow    = "\033[33m"
	ColorCyan      = "\033[36m"
	IconSuccess    = "✔"
	IconError      = "✖"
	IconInfo       = "ℹ"
	EmailHighlight = "\033[1;32m"
	OtherEmails    = "\033[33m"
	MaxWorkers     = 50 // Максимальное количество параллельных обработчиков
)

type FileResult struct {
	Filename   string
	Found      bool
	Records    int
	LogContent string
	Error      error
}

var emailRegex = regexp.MustCompile(`[\w\.=-]+@[\w\.-]+\.[\w]{2,4}`)

func main() {
	folderPath := flag.String("folder", "", "Path to folder with SMTP logs")
	searchEmail := flag.String("email", "", "Email address to search for")
	searchDate := flag.String("date", "", "Date to search (format: YYYY-MM-DD)")
	flag.Parse()

	if *folderPath == "" || *searchEmail == "" {
		log.Fatal("Please provide both folder path and email address")
	}

	// Канал для задач
	jobs := make(chan string, 100)
	// Канал для результатов
	results := make(chan FileResult, 100)
	// Ожидаем завершения всех воркеров
	var wg sync.WaitGroup

	// Запускаем воркеры
	for i := 0; i < MaxWorkers; i++ {
		wg.Add(1)
		go worker(*searchEmail, *searchDate, jobs, results, &wg)
	}

	// Собираем все файлы рекурсивно
	go func() {
		err := filepath.Walk(*folderPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				jobs <- path
			}
			return nil
		})
		if err != nil {
			log.Fatalf("Error walking directory: %v", err)
		}
		close(jobs)
	}()

	// Собираем результаты
	var resultsCollection []FileResult
	go func() {
		for res := range results {
			resultsCollection = append(resultsCollection, res)
		}
	}()

	// Ждем завершения всех задач
	wg.Wait()
	close(results)

	printResults(resultsCollection, *searchEmail)
}

func worker(email, date string, jobs <-chan string, results chan<- FileResult, wg *sync.WaitGroup) {
	defer wg.Done()
	recordStart := regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`)

	for file := range jobs {
		results <- processFile(file, email, date, recordStart)
	}
}

func processFile(filename, email, date string, recordStart *regexp.Regexp) FileResult {
	result := FileResult{Filename: filename}
	file, err := os.Open(filename)
	if err != nil {
		result.Error = err
		return result
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var record strings.Builder
	var found bool
	var totalRecords int

	for scanner.Scan() {
		line := scanner.Text()

		if recordStart.MatchString(line) {
			if record.Len() > 0 {
				totalRecords++
				if checkRecord(email, date, record.String()) {
					found = true
					result.LogContent += record.String() + "\n" + strings.Repeat("─", 40) + "\n"
				}
				record.Reset()
			}
		}
		record.WriteString(line + "\n")
	}

	if record.Len() > 0 {
		totalRecords++
		if checkRecord(email, date, record.String()) {
			found = true
			result.LogContent += record.String() + "\n" + strings.Repeat("─", 40) + "\n"
		}
	}

	result.Found = found
	result.Records = totalRecords
	result.Error = scanner.Err()
	return result
}

func checkRecord(email, date, record string) bool {
	hasEmail := strings.Contains(record, email)

	if date == "" {
		return hasEmail
	}

	hasDate := strings.Contains(record, date)
	return hasEmail && hasDate
}

func highlightEmails(content, searchEmail string) string {
	return emailRegex.ReplaceAllStringFunc(content, func(match string) string {
		if match == searchEmail {
			return EmailHighlight + match + ColorReset
		}
		return OtherEmails + match + ColorReset
	})
}

func printResults(results []FileResult, email string) {
	var foundFiles, errorFiles, notFoundFiles []string
	var totalRecords int

	fmt.Printf("\n%s%s SEARCH RESULTS %s\n",
		ColorCyan,
		strings.Repeat("─", 30),
		ColorReset)

	for _, res := range results {
		totalRecords += res.Records

		if res.Error != nil {
			errorFiles = append(errorFiles,
				fmt.Sprintf("%s[%s]%s %s - %v",
					ColorRed,
					IconError,
					ColorReset,
					res.Filename,
					res.Error))
			continue
		}

		if res.Found {
			foundFiles = append(foundFiles, res.Filename)
			fmt.Printf("\n%s[%s]%s Match found in %s%s%s",
				ColorGreen,
				IconSuccess,
				ColorReset,
				ColorYellow,
				res.Filename,
				ColorReset)
			fmt.Printf("\n%s%s%s\n",
				ColorCyan,
				strings.Repeat("─", 40),
				ColorReset)
			fmt.Println(highlightEmails(res.LogContent, email))
		} else {
			notFoundFiles = append(notFoundFiles, res.Filename)
		}
	}

	fmt.Printf("\n%s%s SUMMARY %s\n",
		ColorCyan,
		strings.Repeat("─", 35),
		ColorReset)

	fmt.Printf("%sTotal files processed:%s %d\n",
		ColorCyan, ColorReset, len(results))

	fmt.Printf("%sFiles with matches:%s    %s%d%s\n",
		ColorCyan, ColorReset,
		ColorGreen, len(foundFiles), ColorReset)

	fmt.Printf("%sFiles without matches:%s %s%d%s\n",
		ColorCyan, ColorReset,
		ColorRed, len(notFoundFiles), ColorReset)

	fmt.Printf("%sTotal records scanned:%s %d\n",
		ColorCyan, ColorReset, totalRecords)

	if len(errorFiles) > 0 {
		fmt.Printf("\n%s%s ERRORS %s\n",
			ColorRed,
			strings.Repeat("─", 35),
			ColorReset)
		for _, err := range errorFiles {
			fmt.Println(err)
		}
	}

	fmt.Printf("\n%sSearch target:%s %s%s%s\n",
		ColorCyan,
		ColorReset,
		ColorGreen,
		email,
		ColorReset)

	fmt.Println(strings.Repeat("═", 50))
}
