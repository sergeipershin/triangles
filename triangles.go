package main

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/fogleman/gg"
)

const minNumTriangles = 4
const maxNumTriangles = 16
const tg30 = 0.57735026918962576450914878050196
const tg30x2 = 1.1547005383792515290182975610039
const scale = 200.0
const indent = 20.0

type triangle struct {
	x int
	y int
	z int
}

func newTriangle(x, y, z int) *triangle {
	return &triangle{
		x: x,
		y: y,
		z: z,
	}
}

func (t *triangle) getCopy() *triangle {
	return newTriangle(t.x, t.y, t.z)
}

func (t *triangle) isEqual(other *triangle) bool {
	return t.x == other.x && t.y == other.y && t.z == other.z
}

func (t *triangle) getCoord(axis int) int {
	switch axis {
	case 1:
		return t.x
	case 2:
		return t.y
	case 3:
		return t.z
	}
	return 0
}

func (t *triangle) getNeighbourCoords(axis int) (int, int, int) {
	var look = t.x + t.y + t.z
	switch axis {
	case 1:
		return t.x, t.y - look, t.z - look
	case 2:
		return t.x - look, t.y, t.z - look
	case 3:
		return t.x - look, t.y - look, t.z
	}
	return 0, 0, 0
}

func (t *triangle) getNeighbour(axis int) *triangle {
	return newTriangle(t.getNeighbourCoords(axis))
}

func (t *triangle) getRotatedCoords(angle int) (int, int, int) {
	switch angle {
	case 1:
		return -t.y, -t.z, -t.x
	case 2:
		return t.z, t.x, t.y
	case 3:
		return -t.x, -t.y, -t.z
	case 4:
		return t.y, t.z, t.x
	case 5:
		return -t.z, -t.x, -t.y
	}
	return t.x, t.y, t.z
}

func (t *triangle) getRotated(angle int) *triangle {
	return newTriangle(t.getRotatedCoords(angle))
}

func (t *triangle) getReflectedCoords(axis int) (int, int, int) {
	switch axis {
	case 1:
		return -t.x, -t.z, -t.y
	case 2:
		return -t.z, -t.y, -t.x
	case 3:
		return -t.y, -t.x, -t.z
	}
	return t.x, t.y, t.z
}

func (t *triangle) getReflected(axis int) *triangle {
	return newTriangle(t.getReflectedCoords(axis))
}

func (t *triangle) getShiftedCoords(shift, axis int) (int, int, int) {
	switch axis {
	case 1:
		return t.x, t.y + shift, t.z - shift
	case 2:
		return t.x - shift, t.y, t.z + shift
	case 3:
		return t.x + shift, t.y - shift, t.z
	}
	return t.x, t.y, t.z
}

func (t *triangle) getShifted(shift, axis int) *triangle {
	return newTriangle(t.getShiftedCoords(shift, axis))
}

func (t *triangle) getCartesianCoords(axis int) (float64, float64, float64, float64) {
	var x1, y1, x2, y2, xf, yf, zf float64
	xf = float64(t.x)
	yf = float64(t.y)
	zf = float64(t.z)
	switch axis {
	case 1:
		x1 = xf / tg30x2
		y1 = xf/2.0 + yf
		x2 = x1
		y2 = -xf/2.0 - zf
	case 2:
		x1 = xf / tg30x2
		y1 = xf/2.0 + yf
		x2 = -(zf + yf) / tg30x2
		y2 = x2*tg30 + yf
	case 3:
		x1 = xf / tg30x2
		y1 = -xf/2.0 - zf
		x2 = -(zf + yf) / tg30x2
		y2 = -x2*tg30 - zf
	}
	return x1, y1, x2, y2
}

type pattern struct {
	triangles   []*triangle
	patternHash string
	validHash   bool
}

func newPattern() *pattern {
	return &pattern{
		triangles: make([]*triangle, 0, maxNumTriangles),
	}
}

func (p *pattern) validateHash() {
	var tstr string
	arr := make([]string, 0, maxNumTriangles)
	if !p.validHash {
		for i := 0; i < len(p.triangles); i++ {
			tstr = fmt.Sprintf("%d,%d,%d", p.triangles[i].x, p.triangles[i].y, p.triangles[i].z)
			arr = append(arr, tstr)
		}
		sort.Strings(arr)
		p.patternHash = strings.Join(arr, " ")
		p.validHash = true
	}
}

func (p *pattern) getCopy() *pattern {
	pCopy := newPattern()
	for i := 0; i < len(p.triangles); i++ {
		pCopy.addTriangle(p.triangles[i].getCopy())
	}
	pCopy.validateHash()
	return pCopy
}

func (p *pattern) len() int {
	return len(p.triangles)
}

func (p *pattern) isEqual(other *pattern) bool {
	var pAligned, otherRotated, otherAligned *pattern
	var /*patternsAreEqual,*/ foundEqualPattern bool
	freeAxis := 3
	foundEqualPattern = false
	pAligned = p.getAligned(freeAxis)
	pAligned.validateHash()
	otherRotated = other
	for i := 1; i <= 6; i++ {
		for j := 1; j <= 2; j++ {
			if j == 1 {
				otherAligned = otherRotated.getAligned(freeAxis)
			} else {
				otherAligned = otherRotated.getReflected(freeAxis).getAligned(freeAxis)
			}
			otherAligned.validateHash()
			if pAligned.patternHash == otherAligned.patternHash {
				foundEqualPattern = true
				break
			}
		}
		if foundEqualPattern {
			break
		}
		if i < 6 {
			otherRotated = otherRotated.getRotated(1)
		}
	}
	return foundEqualPattern
}

func (p *pattern) contains(t *triangle) bool {
	result := false
	for i := 0; i < len(p.triangles); i++ {
		if p.triangles[i].isEqual(t) {
			result = true
			break
		}
	}
	return result
}

func (p *pattern) addTriangle(t *triangle) {
	p.triangles = append(p.triangles, t)
	p.validHash = false
}

func (p *pattern) getMinCoord(axis int) int {
	numTriangles := len(p.triangles)
	if numTriangles == 0 {
		return 0
	}
	var minCoord = p.triangles[0].getCoord(axis)
	for i := 1; i < len(p.triangles); i++ {
		currCoord := p.triangles[i].getCoord(axis)
		if currCoord < minCoord {
			minCoord = currCoord
		}
	}
	return minCoord
}

func (p *pattern) getMaxCoord(axis int) int {
	numTriangles := len(p.triangles)
	if numTriangles == 0 {
		return 0
	}
	var maxCoord = p.triangles[0].getCoord(axis)
	for i := 1; i < len(p.triangles); i++ {
		currCoord := p.triangles[i].getCoord(axis)
		if currCoord > maxCoord {
			maxCoord = currCoord
		}
	}
	return maxCoord
}

func (p *pattern) getShifted(shift, axis int) *pattern {
	shifted := newPattern()
	for i := 0; i < len(p.triangles); i++ {
		shifted.addTriangle(p.triangles[i].getShifted(shift, axis))
	}
	return shifted
}

func (p *pattern) getRotated(angle int) *pattern {
	rotated := newPattern()
	for i := 0; i < len(p.triangles); i++ {
		rotated.addTriangle(p.triangles[i].getRotated(angle))
	}
	return rotated
}

func (p *pattern) getReflected(axis int) *pattern {
	reflected := newPattern()
	for i := 0; i < len(p.triangles); i++ {
		reflected.addTriangle(p.triangles[i].getReflected(axis))
	}
	return reflected
}

func (p *pattern) getAligned(freeAxis int) *pattern {
	var aligned *pattern
	var min_coord, max_coord int
	switch freeAxis {
	case 1:
		max_coord = p.getMaxCoord(2)
		aligned = p.getShifted(max_coord, 3)
		min_coord = aligned.getMinCoord(3)
		aligned = aligned.getShifted(-min_coord, 2)
	case 2:
		max_coord = p.getMaxCoord(3)
		aligned = p.getShifted(max_coord, 1)
		min_coord = aligned.getMinCoord(1)
		aligned = aligned.getShifted(-min_coord, 3)
	case 3:
		max_coord = p.getMaxCoord(1)
		aligned = p.getShifted(max_coord, 2)
		min_coord = aligned.getMinCoord(2)
		aligned = aligned.getShifted(-min_coord, 1)
	}
	return aligned
}

func (p *pattern) getCentered() *pattern {
	var centered *pattern
	var min_coord, max_coord, mean_coord int
	min_coord = p.getMinCoord(1)
	max_coord = p.getMaxCoord(1)
	mean_coord = (min_coord + max_coord) / 2
	centered = p.getShifted(mean_coord, 2)
	min_coord = centered.getMinCoord(2)
	max_coord = centered.getMaxCoord(2)
	mean_coord = (min_coord + max_coord) / 2
	centered = centered.getShifted(-mean_coord, 1)
	return centered
}

type line struct {
	x1, y1, x2, y2 float64
	bold           bool
}

func newLine(x1, y1, x2, y2 float64, bold bool) line {
	return line{
		x1:   x1,
		y1:   y1,
		x2:   x2,
		y2:   y2,
		bold: bold,
	}
}

type patternImage struct {
	xMin, yMin, xMax, yMax float64
	width                  float64
	height                 float64
	img                    *gg.Context
}

func newPatternImage() patternImage {
	return patternImage{}
}

func (pimg *patternImage) toReal(x, y float64) (float64, float64) {
	return x*scale + pimg.width/2, pimg.height/2 - y*scale
}

func (pimg *patternImage) drawPattern(p *pattern) {
	var x, y, x0, y0, x1, y1, x2, y2, x3, y3, x4, radius float64
	var t, tn *triangle
	var l line
	lines := make([]line, 0, maxNumTriangles*3)
	radius = 0.0
	for i := 0; i < len(p.triangles); i++ {
		t = p.triangles[i]
		for axis := 1; axis <= 3; axis++ {
			x1, y1, x2, y2 = t.getCartesianCoords(axis)
			radius = max(math.Abs(x1), math.Abs(y1), math.Abs(x2), math.Abs(y2), radius)
			tn = t.getNeighbour(axis)
			if p.contains(tn) {
				l = newLine(x1, y1, x2, y2, false)
			} else {
				l = newLine(x1, y1, x2, y2, true)
			}
			lines = append(lines, l)
		}
	}
	pimg.xMin = -radius - 1
	pimg.yMin = pimg.xMin
	pimg.xMax = -pimg.xMin
	pimg.yMax = pimg.xMax
	pimg.width = (pimg.xMax-pimg.xMin)*scale + indent
	pimg.height = (pimg.yMax-pimg.yMin)*scale + indent

	pimg.img = gg.NewContext(int(pimg.width), int(pimg.height))
	pimg.img.SetRGB(1, 1, 1) // белый фон
	pimg.img.Clear()

	for x = math.Round(pimg.xMin * tg30x2); x <= pimg.xMax*tg30x2; x++ {
		x1, y1 = pimg.toReal(x/tg30x2, pimg.yMin)
		x2, y2 = pimg.toReal(x/tg30x2, pimg.yMax)
		pimg.img.SetRGB(0.002, 0.002, 0.002)
		pimg.img.SetLineWidth(0.3)
		pimg.img.DrawLine(x1, y1, x2, y2)
		pimg.img.Stroke()
	}
	for y = math.Round(pimg.yMax - pimg.xMin*tg30); y >= pimg.yMin-pimg.xMax*tg30; y-- {
		x1 = pimg.xMin
		y1 = y + pimg.xMin*tg30
		x2 = pimg.xMax
		y2 = y1 + (pimg.xMax-pimg.xMin)*tg30
		if y1 < pimg.yMin {
			x1 = pimg.xMin + (pimg.yMin-y1)/tg30
			y1 = pimg.yMin
		}
		if y2 > pimg.yMax {
			x2 = pimg.xMax - (y2-pimg.yMax)/tg30
			y2 = pimg.yMax
		}
		x3 = pimg.xMax - x1 + pimg.xMin
		x4 = pimg.xMax - x2 + pimg.xMin
		x1, y1 = pimg.toReal(x1, y1)
		x2, y2 = pimg.toReal(x2, y2)
		x3, _ = pimg.toReal(x3, y1)
		x4, _ = pimg.toReal(x4, y2)
		pimg.img.SetRGB(0.002, 0.002, 0.002)
		pimg.img.SetLineWidth(0.3)
		pimg.img.DrawLine(x1, y1, x2, y2)
		pimg.img.Stroke()
		pimg.img.DrawLine(x3, y1, x4, y2)
		pimg.img.Stroke()
	}

	x0 = 0
	y0 = 0
	x1 = 0
	y1 = pimg.yMax
	x2 = pimg.xMin
	y2 = pimg.xMin * tg30
	x3 = pimg.xMax
	y3 = -pimg.xMax * tg30
	x0, y0 = pimg.toReal(x0, y0)
	x1, y1 = pimg.toReal(x1, y1)
	x2, y2 = pimg.toReal(x2, y2)
	x3, y3 = pimg.toReal(x3, y3)
	pimg.img.SetRGB(0.04, 0.04, 0.04)
	pimg.img.SetLineWidth(1)
	pimg.img.DrawLine(x0, y0, x1, y1)
	pimg.img.Stroke()
	pimg.img.DrawLine(x0, y0, x2, y2)
	pimg.img.Stroke()
	pimg.img.DrawLine(x0, y0, x3, y3)
	pimg.img.Stroke()

	for i := 0; i < len(lines); i++ {
		l = lines[i]
		x1, y1 = pimg.toReal(l.x1, l.y1)
		x2, y2 = pimg.toReal(l.x2, l.y2)
		pimg.img.SetRGB(0.0, 0.0, 0.0)
		if l.bold {
			pimg.img.SetLineWidth(5)
		} else {
			pimg.img.SetLineWidth(2)
		}
		pimg.img.DrawLine(x1, y1, x2, y2)
		pimg.img.Stroke()
	}
}

func (pimg *patternImage) saveAsPNG(path string) {
	pimg.img.SavePNG(path)
}

type patternsCollection struct {
	patterns []*pattern
}

func newPatternsCollection() *patternsCollection {
	return &patternsCollection{
		patterns: make([]*pattern, 0, maxNumTriangles*maxNumTriangles),
	}
}

func (pc *patternsCollection) generatePatterns(toAdd int, sketch *pattern) {
	var neighbour *triangle
	var newSketch *pattern
	var foundNewPattern bool
	if sketch.len() == 0 {
		sketch.addTriangle(newTriangle(0, 1, 0))
		if toAdd > 1 {
			pc.generatePatterns(toAdd-1, sketch)
		} else {
			pc.patterns = append(pc.patterns, sketch)
		}
		return
	} else if sketch.len() <= 2 {
		neighbour = sketch.triangles[0].getNeighbour(sketch.len())
		newSketch = sketch.getCopy()
		newSketch.addTriangle(neighbour)
		if toAdd > 1 {
			pc.generatePatterns(toAdd-1, newSketch)
		} else {
			foundNewPattern = true
			for j := 0; j < len(pc.patterns); j++ {
				if pc.patterns[j].isEqual(newSketch) {
					foundNewPattern = false
					break
				}
			}
			if foundNewPattern {
				pc.patterns = append(pc.patterns, newSketch.getCentered())
			}
		}
	} else {
		for i := 0; i < sketch.len(); i++ {
			for axis := 1; axis <= 3; axis++ {
				neighbour = sketch.triangles[i].getNeighbour(axis)
				if sketch.contains(neighbour) {
					continue
				}
				newSketch = sketch.getCopy()
				newSketch.addTriangle(neighbour)
				if toAdd > 1 {
					pc.generatePatterns(toAdd-1, newSketch)
				} else {
					foundNewPattern = true
					for j := 0; j < len(pc.patterns); j++ {
						if pc.patterns[j].isEqual(newSketch) {
							foundNewPattern = false
							break
						}
					}
					if foundNewPattern {
						pc.patterns = append(pc.patterns, newSketch.getCentered())
					}
				}
			}
		}
	}
}

func main() {
	var pimg patternImage
	var numTriangles int

	fmt.Printf("Введите количество треугольников (%d-%d): ", minNumTriangles, maxNumTriangles)
	fmt.Scanf("%d", &numTriangles)
	if numTriangles < minNumTriangles || numTriangles > maxNumTriangles {
		fmt.Print("Неправильное значение")
		return
	}

	os.Mkdir(fmt.Sprintf("%d", numTriangles), 0755)
	pattCol := newPatternsCollection()
	sk := newPattern()
	pattCol.generatePatterns(numTriangles, sk)
	for i := 0; i < len(pattCol.patterns); i++ {
		pimg = newPatternImage()
		pimg.drawPattern(pattCol.patterns[i])
		pimg.saveAsPNG(fmt.Sprintf("%d/%d.png", numTriangles, i))
	}
}
