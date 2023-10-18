package wkttoorb

import (
	"fmt"
	"strconv"

	"github.com/paulmach/orb"
	"github.com/pkg/errors"
)

type Parser struct {
	*Lexer
}

func (p *Parser) Parse() (orb.Geometry, error) {
	t, err := p.scanToken()
	if err != nil {
		return nil, err
	}
	switch t.ttype {
	case Point:
		return p.parsePoint()
	case Linestring:
		return p.parseLineString()
	case Polygon:
		return p.parsePolygon()
	case Multipoint:
		line, err := p.parseLineString()
		return orb.MultiPoint(line), err
	case MultilineString:
		poly, err := p.parsePolygon()
		multiline := make(orb.MultiLineString, 0, len(poly))
		for _, ring := range poly {
			multiline = append(multiline, orb.LineString(ring))
		}
		return multiline, err
	case MultiPolygon:
		return p.parseMultiPolygon()
	default:
		return nil, fmt.Errorf("Parse unexpected token %s on pos %d expected geometry type", t.lexeme, t.pos)
	}
}

func (p *Parser) parsePoint() (point orb.Point, err error) {
	t, err := p.scanToken()
	if err != nil {
		return point, err
	}
	switch t.ttype {
	case Empty:
		point = orb.Point{0, 0}
	case Z, M, ZM:
		t1, err := p.scanToken()
		if err != nil {
			return point, err
		}
		if t1.ttype == Empty {
			point = orb.Point{0, 0}
			break
		}
		if t1.ttype != LeftParen {
			return point, fmt.Errorf("parse point unexpected token %s on pos %d", t.lexeme, t.pos)
		}
		fallthrough
	case LeftParen:
		switch t.ttype { // reswitch on the type because of the fallthrough
		case Z, M:
			point, err = p.parseCoordDrop1()
		case ZM:
			point, err = p.parseCoordDrop2()
		default:
			point, err = p.parseCoord()
		}
		if err != nil {
			return point, err
		}

		t, err = p.scanToken()
		if err != nil {
			return point, err
		}

		if t.ttype != RightParen {
			return point, fmt.Errorf("parse point unexpected token %s on pos %d expected )", t.lexeme, t.pos)
		}
	default:
		return point, fmt.Errorf("parse point unexpected token %s on pos %d", t.lexeme, t.pos)
	}

	t, err = p.scanToken()
	if err != nil {
		return point, err
	}
	if t.ttype != Eof {
		return point, fmt.Errorf("parse point unexpected token %s on pos %d", t.lexeme, t.pos)
	}

	return point, nil
}

func (p *Parser) parseLineString() (line orb.LineString, err error) {
	line = make([]orb.Point, 0)
	t, err := p.scanToken()
	if err != nil {
		return nil, err
	}
	switch t.ttype {
	case Empty:
	case Z, M, ZM:
		t1, err := p.scanToken()
		if err != nil {
			return line, err
		}
		if t1.ttype == Empty {
			break
		}
		if t1.ttype != LeftParen {
			return line, fmt.Errorf("unexpected token %s on pos %d expected '('", t.lexeme, t.pos)
		}
		fallthrough
	case LeftParen:
		line, err = p.parseLineStringText(t.ttype)
		if err != nil {
			return line, err
		}
	default:
		return line, fmt.Errorf("unexpected token %s on pos %d", t.lexeme, t.pos)
	}

	t, err = p.scanToken()
	if err != nil {
		return line, err
	}
	if t.ttype != Eof {
		return line, fmt.Errorf("unexpected token %s on pos %d, expected Eof", t.lexeme, t.pos)
	}

	return line, nil
}

func (p *Parser) parseLineStringText(ttype tokenType) (line orb.LineString, err error) {
	line = make([]orb.Point, 0)
	for {
		var point orb.Point
		switch ttype {
		case Z, M:
			point, err = p.parseCoordDrop1()
		case ZM:
			point, err = p.parseCoordDrop2()
		default:
			point, err = p.parseCoord()
		}
		if err != nil {
			return line, err
		}
		line = append(line, point)
		t, err := p.scanToken()
		if err != nil {
			return line, err
		}
		if t.ttype == RightParen {
			break
		} else if t.ttype != Comma {
			return line, fmt.Errorf("unexpected token %s on pos %d expected ','", t.lexeme, t.pos)
		}
	}
	return line, nil
}

func (p *Parser) parsePolygon() (poly orb.Polygon, err error) {
	poly = make([]orb.Ring, 0)
	t, err := p.scanToken()
	if err != nil {
		return poly, err
	}
	switch t.ttype {
	case Empty:
	case Z, M, ZM:
		t1, err := p.scanToken()
		if err != nil {
			return poly, err
		}
		if t1.ttype == Empty {
			break
		}
		if t1.ttype != LeftParen {
			return poly, fmt.Errorf("unexpected token %s on pos %d expected '('", t.lexeme, t.pos)
		}
		fallthrough
	case LeftParen:
		poly, err = p.parsePolygonText(t.ttype)
		if err != nil {
			return poly, err
		}
	default:
		return poly, fmt.Errorf("unexpected token %s on pos %d", t.lexeme, t.pos)
	}

	t, err = p.scanToken()
	if err != nil {
		return poly, err
	}
	if t.ttype != Eof {
		return poly, fmt.Errorf("unexpected token %s on pos %d, expected Eof", t.lexeme, t.pos)
	}

	return poly, nil
}

func (p *Parser) parsePolygonText(ttype tokenType) (poly orb.Polygon, err error) {
	poly = make([]orb.Ring, 0)
	for {
		var line orb.LineString
		t, err := p.scanToken()
		if err != nil {
			return poly, err
		}
		if t.ttype != LeftParen {
			return poly, fmt.Errorf("unexpected token %s on pos %d expected '('", t.lexeme, t.pos)
		}
		line, err = p.parseLineStringText(ttype)
		if err != nil {
			return poly, err
		}
		poly = append(poly, orb.Ring(line))
		t, err = p.scanToken()
		if err != nil {
			return poly, err
		}
		if t.ttype == RightParen {
			break
		} else if t.ttype != Comma {
			return poly, fmt.Errorf("unexpected token %s on pos %d expected ','", t.lexeme, t.pos)
		}
	}
	return poly, nil
}

func (p *Parser) parseMultiPolygon() (multi orb.MultiPolygon, err error) {
	multi = make([]orb.Polygon, 0)
	t, err := p.scanToken()
	if err != nil {
		return multi, err
	}
	switch t.ttype {
	case Empty:
	case Z, M, ZM:
		t1, err := p.scanToken()
		if err != nil {
			return multi, err
		}
		if t1.ttype == Empty {
			break
		}
		if t1.ttype != LeftParen {
			return multi, fmt.Errorf("unexpected token %s on pos %d expected '('", t.lexeme, t.pos)
		}
		fallthrough
	case LeftParen:
		multi, err = p.parseMultiPolygonText(t.ttype)
		if err != nil {
			return multi, err
		}
	default:
		return multi, fmt.Errorf("unexpected token %s on pos %d", t.lexeme, t.pos)
	}

	t, err = p.scanToken()
	if err != nil {
		return multi, err
	}
	if t.ttype != Eof {
		return multi, fmt.Errorf("unexpected token %s on pos %d, expected Eof", t.lexeme, t.pos)
	}

	return multi, nil
}

func (p *Parser) parseMultiPolygonText(ttype tokenType) (multi orb.MultiPolygon, err error) {
	multi = make([]orb.Polygon, 0)
	for {
		var poly orb.Polygon
		t, err := p.scanToken()
		if err != nil {
			return multi, err
		}
		if t.ttype != LeftParen {
			return multi, fmt.Errorf("unexpected token %s on pos %d expected '('", t.lexeme, t.pos)
		}
		poly, err = p.parsePolygonText(ttype)
		if err != nil {
			return multi, err
		}
		multi = append(multi, poly)
		t, err = p.scanToken()
		if err != nil {
			return multi, err
		}
		if t.ttype == RightParen {
			break
		} else if t.ttype != Comma {
			return multi, fmt.Errorf("unexpected token %s on pos %d expected ','", t.lexeme, t.pos)
		}
	}
	return multi, nil
}

func (p *Parser) parseCoord() (point orb.Point, err error) {
	t1, err := p.scanToken()
	if err != nil {
		return point, err
	}
	if t1.ttype != Float {
		return point, fmt.Errorf("parse coordinates unexpected token %s on pos %d", t1.lexeme, t1.pos)
	}
	t2, err := p.scanToken()
	if err != nil {
		return point, err
	}
	if t2.ttype != Float {
		return point, fmt.Errorf("parse coordinates unexpected token %s on pos %d", t1.lexeme, t2.pos)
	}

	c1, err := strconv.ParseFloat(t1.lexeme, 64)
	if err != nil {
		return point, errors.Wrap(err, fmt.Sprintf("invalid lexeme %s for token on pos %d", t1.lexeme, t1.pos))
	}
	c2, err := strconv.ParseFloat(t2.lexeme, 64)
	if err != nil {
		return point, errors.Wrap(err, fmt.Sprintf("invalid lexeme %s for token on pos %d", t2.lexeme, t2.pos))
	}

	return orb.Point{c1, c2}, nil
}

func (p *Parser) parseCoordDrop1() (point orb.Point, err error) {
	point, err = p.parseCoord()
	if err != nil {
		return point, err
	}

	// drop the last value Z or M coordinates are not really supported
	t, err := p.scanToken()
	if err != nil {
		return point, err
	}
	if t.ttype != Float {
		return point, fmt.Errorf("parseCoordDrop1 unexpected token %s on pos %d expected Float", t.lexeme, t.pos)
	}

	return point, nil
}

func (p *Parser) parseCoordDrop2() (point orb.Point, err error) {
	point, err = p.parseCoord()
	if err != nil {
		return point, err
	}

	// drop the last value M values
	// and Z coordinates are not really supported
	for i := 0; i < 2; i++ {
		t, err := p.scanToken()
		if err != nil {
			return point, err
		}
		if t.ttype != Float {
			return point, fmt.Errorf("parseCoordDrop2 unexpected token %s on pos %d expected Float", t.lexeme, t.pos)
		}
	}

	return point, nil
}
