package fetcher

import (
	"context"
	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type LibDescription struct {
	_          tlb.Magic        `tlb:"$00"`
	Lib        *cell.Cell       `tlb:"^"`
	Publishers *cell.Dictionary `tlb:"dict inline 256"`
}

func (s *Service) GetAccountLibraries(ctx context.Context, raw *tlb.Account) (*cell.Cell, error) {
	hashes, err := findLibraries(raw.Code)

	if err != nil {
		return nil, errors.Wrapf(err, "find libraries")
	}

	libs, err := s.API.GetLibraries(ctx, hashes...)

	if err != nil {
		return nil, errors.Wrapf(err, "get libraries")
	}

	libsMap := cell.NewDict(256)

	for i, hash := range hashes {
		desc := LibDescription{Lib: libs[i]}

		h := cell.BeginCell().MustStoreSlice(hash, 256).EndCell()
		t, err := tlb.ToCell(desc)

		if err != nil {
			return nil, err
		}

		err = libsMap.Set(h, t)
		s.libraries.set(hash, &desc)

		if err != nil {
			return nil, err
		}
	}

	return libsMap.ToCell()
}

func getLibraryHash(code *cell.Cell) ([]byte, error) {
	hash, err := code.BeginParse().LoadBinarySnake()

	if err != nil {
		return nil, err
	}

	return hash[1:], nil
}

// TODO recursive Refs
func findLibraries(code *cell.Cell) ([][]byte, error) {
	hashes := make([][]byte, 0)

	if code.GetType() == cell.LibraryCellType {
		hash, err := getLibraryHash(code)

		if err != nil {
			return nil, err
		}

		hashes = append(hashes, hash)
	}

	return hashes, nil
}
