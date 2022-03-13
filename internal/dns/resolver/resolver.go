package resolver

import (
	"errors"
	"log"

	"github.com/bluguard/dnshield/internal/dns/dto"
)

type Resolver interface {
	Resolve(dto.Question) (dto.Record, bool)
	Name() string
}

//Resolver
type ResolverChain struct {
	chain []Resolver
}

func (resolverChain *ResolverChain) Resolve(message dto.Message) dto.Message {
	records := resolverChain.resolveAll(message.Question)
	response := dto.Message{
		ID:            message.ID, //TODO check the method to set the  id of a response
		Header:        dto.STANDARD_RESPONSE,
		QuestionCount: message.QuestionCount,
		ResponseCount: uint16(len(records)),
		Question:      message.Question,
		Response:      records,
	}

	return response
}

func (resolverChain *ResolverChain) resolveAll(questions []dto.Question) []dto.Record {
	records := make([]dto.Record, 0, 4)
	for _, question := range questions {
		r, err := resolverChain.resolveOne(question)
		if err != nil {
			log.Println(err.Error())
		} else {
			records = append(records, r)
		}
	}
	return records
}

func (resolverChain *ResolverChain) resolveOne(question dto.Question) (dto.Record, error) {
	for _, resolver := range resolverChain.chain {
		if record, ok := resolver.Resolve(question); ok {
			return record, nil
		}
	}
	return dto.Record{}, errors.New("no record found for " + question.Name)
}
