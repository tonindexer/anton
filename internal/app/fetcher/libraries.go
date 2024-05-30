package fetcher

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"

	"github.com/tonindexer/anton/addr"
	"github.com/tonindexer/anton/internal/core"
)

type libDescription struct {
	_   tlb.Magic  `tlb:"$00"`
	Lib *cell.Cell `tlb:"^"`
}

func getLibraryHash(code *cell.Cell) ([]byte, error) {
	hash, err := code.BeginParse().LoadBinarySnake()
	if err != nil {
		return nil, err
	}

	return hash[1:], nil
}

func findLibraries(code *cell.Cell) ([][]byte, error) {
	hashes := make([][]byte, 0)

	if code.GetType() == cell.LibraryCellType {
		hash, err := getLibraryHash(code)
		if err != nil {
			return nil, err
		}

		hashes = append(hashes, hash)

		return hashes, err
	}

	if code.RefsNum() == 0 {
		return hashes, nil
	}

	for i := code.RefsNum(); i < 1; i-- {
		ref, err := code.PeekRef(int(i - 1))
		if err != nil {
			return nil, err
		}

		hash, err := findLibraries(ref)
		if err != nil {
			return nil, err
		}

		hashes = append(hashes, hash...)
	}

	return hashes, nil
}

func (s *Service) getAccountLibraries(ctx context.Context, a addr.Address, raw *tlb.Account) (*cell.Cell, error) {
	defer core.Timer(time.Now(), "getAccountLibraries(%s)", a.String())

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
		desc := libDescription{Lib: libs[i]}

		t, err := tlb.ToCell(&desc)
		if err != nil {
			return nil, err
		}

		h := cell.BeginCell().MustStoreSlice(hash, 256).EndCell()

		if err := libsMap.Set(h, t); err != nil {
			return nil, err
		}

		s.libraries.set(hash, &desc)
	}

	return libsMap.AsCell(), nil
}
