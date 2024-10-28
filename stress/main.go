package main

var authHeader = "Bearer 5Mxrg3TkCRq4aMy4PyO8QYA7BiWUqHy9fPlVbSruAlDpGj10ry4mgbbetL79M12S"
var baseUrl = "https://test.bible.clementineleaf.top/stress"

var testNames = []string{"noindex", "tsv", "createatuser"}

func main() {

	for _, name := range testNames {
		run_test(name)
	}
}

func run_test(name string) {

}
