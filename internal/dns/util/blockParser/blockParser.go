package blockparser

import (
	"bufio"
	"net/http"
	"strings"

	"github.com/bluguard/dnshield/internal/dns/client/blocker"
)

const (
	valideLineStart = "0.0.0.0"
	commentStart    = "#"
	valueSeparator  = " "
)

type BlockParser struct {
	Url string
}

var _ blocker.Initializer = (&BlockParser{}).Feed

func (p *BlockParser) Feed(add func(name string)) {
	resp, err := http.Get(p.Url)
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		text := scanner.Text()
		if !strings.HasPrefix(text, valideLineStart) {
			continue
		}
		text = strings.Split(text, commentStart)[0]
		if !strings.Contains(text, " ") {
			continue
		}
		text = strings.Split(text, valueSeparator)[1]
		add(text)
	}
}
