package main

import (
	"runtime"
	"github.com/go-gl/glfw/v3.4/glfw"
	"github.com/go-gl/gl/v2.1/gl"
	"fmt"
	"os"
	"bufio"
	"errors"
	"strings"
	"unicode"
	"strconv"
	"slices"
)

type point struct {
	x,y,z float32
}
type line struct {
	n,m point
}
type triangle struct {
	a,b,c point
}
type variable interface {
	astriangle()( triangle,error)
	asline()( line,error)
	aspoint()( point,error)
	typeof() string
}
func (t triangle)typeof() string {
	return "triangle"	
}
func (t triangle)asline() (line,error) {
	return line{},errors.New("cannot get line from triangle")
}
func (t triangle)aspoint() (point,error) {
	return point{},errors.New("cannot get point from triangle")
}
func(t triangle)astriangle() (triangle,error) {
	return t,nil
}
func (l line)typeof() string {
	return "line"	
}
func (l line)asline() (line,error) {
	return l,nil
}
func (l line)aspoint() (point,error) {
	return point{},errors.New("cannot get point from line")
}
func(l line)astriangle() (triangle,error) {
	return triangle{},errors.New("cannot get triangle from line")
}
func (p point)typeof() string {
	return "point"	
}
func (p point)asline() (line,error) {
	return line{},errors.New("cannot get line from point")
}
func (p point)aspoint() (point,error) {
	return p,nil
}
func(p point)astriangle() (triangle,error) {
	return triangle{},errors.New("cannot get triangle from point")
}
func display(v variable) {
	gl.PointSize(10)
	if v.typeof() == "point"{
		gl.Begin(gl.POINTS)
		gl.Color3f(1,0,0)

		p,err := v.aspoint()
		if err != nil {
			panic(err)
		}
		gl.Vertex3f(p.x,p.y,p.z)
		gl.End()

	} else if v.typeof() == "triangle" {

		t,err := v.astriangle()
		if err != nil {
			panic(err)
		}
		gl.Begin(gl.TRIANGLES)
			gl.Color3f(1,1,1)
			gl.Vertex3f(t.a.x,t.a.y,t.a.z)
			gl.Vertex3f(t.b.x,t.b.y,t.b.z)
			gl.Vertex3f(t.c.x,t.c.y,t.c.z)
		gl.End()
		//	display(t.a)
		//	display(t.b)
		//	display(t.c)
	
	} else if v.typeof() == "line" {
		gl.Begin(gl.LINES)
		gl.Color3f(1,1,0)

		l ,err := v.asline()
		if err != nil {
			panic(err)
		}
		gl.Vertex3f(l.n.x,l.n.y,l.n.z)
		gl.Vertex3f(l.m.x,l.m.y,l.m.z)
		gl.End()

		
	} else {
		panic(fmt.Sprintf("%s unknown type in display",v.typeof()))
	}

}
func splitAssumingConstructs(this string) ([]string,error) {
	var returned []string
	for this != "" {
		i := strings.IndexFunc(this,func(c rune) bool {return c == '[' || c == ','} )
		if i == -1 {
			returned = append(returned,this)
			return returned,nil
		}
	
		if this[i] == '[' {	
			var sp int
			j := strings.IndexFunc(this[i:],func(c rune) bool {
				if c == '[' {
					sp++
				} else if c == ']' {
					sp--
				}
				if sp == 0 {
					return true
				}
				return false
			})
			if j == -1 {
				return returned,errors.New("stray [")
			}
			returned = append(returned,this[:i+1+j])
			this = this[i+1+j:]
			if this != "" {
				if this[0] == ',' {
					this = this[1:]
				}
			}
		} else if this[i] == ',' {
			returned = append(returned,this[:i])
			this = this[i:]
			this = this[1:]
		}
		
	}
	return returned,nil
	
}
const special = "|\\+=-(){}*&^%$#@!~\".><?/"
type variables map[string]variable
func (context variables)valueof(this string) (variable,error){
	strings.ReplaceAll(this," ","")
	if this == "" {
		return nil,errors.New("nil expr some how got into valueof")
	}
	var allletters bool = true
	for i := range this {
		if !unicode.IsLetter(rune(this[i])){
			allletters = false	
		}
	}
	if allletters {
		v,ok := context[this]
		if !ok {
			return nil,errors.New("variable does not exist")
		}
		return v,nil
	}
	if this[0] != '[' {
		return nil,errors.New("grouped with no [")
	}
 	if this[len(this)-1] != ']' {
		return nil,errors.New("grouped but no ]")
	}
	fmt.Println(this)

	things,err := splitAssumingConstructs(this[1:len(this)-1])
	if err != nil {
		return nil,err
	}
	
	if len(things) == 3 && slices.ContainsFunc(things,func(s string) bool {return unicode.IsDigit(rune(s[0]))}) {
		var floats [3]float32
		var collective error
		for i := range things {
			f,err := strconv.ParseFloat(things[i],32)
			floats[i] = float32(f)
			collective = errors.Join(collective,err)
		}
		if collective != nil {
			return nil,collective
		}
		return point{floats[0],floats[1],floats[2]},nil
	} else if len(things) == 2 {
		var collective error
		var l line
		var v[2] variable
		var err error
		v[0],err = context.valueof(things[0])
		collective = errors.Join(collective,err)
 		v[1],err = context.valueof(things[1])
		collective = errors.Join(collective,err)
		l.n ,err = v[0].aspoint()
		collective = errors.Join(collective,err)
		l.m ,err = v[1].aspoint()
		collective = errors.Join(collective,err)
		if collective != nil {
			return nil,collective
		}
		return l,nil
	}else if len(things) == 3 {
		var collective error
		var t triangle
		var v [3]variable
		var err error
		v[0],err = context.valueof(things[0])
		collective = errors.Join(collective,err)
		collective = errors.Join(collective,err)
		v[1],err = context.valueof(things[1])
		collective = errors.Join(collective,err)
		collective = errors.Join(collective,err)
		v[2],err = context.valueof(things[2])
		collective = errors.Join(collective,err)
		if collective != nil {
			return nil,collective
		}
		t.a,err = v[0].aspoint()
		t.b,err = v[1].aspoint()
		t.c,err = v[2].aspoint()


		return t,nil	
	} else {
		return nil,errors.New(fmt.Sprintf("this is 3d program not %dd program",len(things)))
	}
	return nil,nil;
}
func main() {
	runtime.LockOSThread()
	verts := make(variables)
	input := make(chan string)
	stdin := bufio.NewScanner(os.Stdin)
	go func() {
		for stdin.Scan() {
			input <- stdin.Text()
			
		} 
	}()
	err := glfw.Init()
	if err != nil {
		panic(err)
	}
	defer glfw.Terminate()

	window, err := glfw.CreateWindow(640, 480, "this", nil, nil)
	if err != nil {
		panic(err)
	}
	err = gl.Init() 
	if err != nil {
		panic(err)
	}
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.DepthMask(true)
	window.MakeContextCurrent()
	glfw.SwapInterval(1)
	var names []string
	for !window.ShouldClose() {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		gl.LoadIdentity()
		
		for _,i := range names {
			display(verts[i])
		}
		
		select {
			case stuff := <- input:
				if stuff == "" {
					continue
				}
				var typename,left,right string
 				if strings.ReplaceAll(stuff," ","") == "triangle" {
					var i int
					var ok bool 
					for ok {
						name := "t"+strconv.Itoa(i)
						_,ok = verts[name]
						i++
					}
					stuff = fmt.Sprintf("triangle t%d = [[0,0.5,0.1],[-0.5,-0.5,0.1],[0.5,-0.5,0.1]]",len(verts))
				}
				if strings.ReplaceAll(stuff," ","") == "point" {
					var i int
					var ok bool 
					for ok {
						name := "p"+strconv.Itoa(i)
						_,ok = verts[name]
						i++
					}
					stuff = fmt.Sprintf("point p%d = [0,0,0]",len(verts))
				}
				if strings.ReplaceAll(stuff," ","") == "line" {
					var i int
					var ok bool 
					for ok {
						name := "l"+strconv.Itoa(i)
						_,ok = verts[name]
						i++
					}
					stuff = fmt.Sprintf("line l%d = [.5,0,0.2],[-.5,0,0.2]",len(verts))
				}
				i := strings.Index(stuff,"=")
				if i != -1 {
					left = stuff[:i]
					right = stuff[i+1:]
				} else {
					left = stuff
				}
				i = strings.Index(left," ")
				if i == -1 {
					stuff = strings.ReplaceAll(left," ","")
					v,ok := verts[left]
					if right == "" {
						fmt.Println(v)
					} else {
						if !ok {
							fmt.Println(left+":","undeclared")
						}
						
						v,err2 := verts.valueof(right)
						if err2 != nil {
							fmt.Println(right+":",err)
							continue
						}
						var v2 variable
						switch(verts[left].typeof()){
							case "triangle":
								v2,err = v.astriangle()
							case "point":
								v2,err = v.aspoint()
							case "line":
		 						v2,err = v.asline()
		
						}
						fmt.Println(v2,v)
						if err != nil {
							fmt.Println(right+":",err)
							continue
						}
						verts[left] = v2
					}
				} else {
					typename = left[:i]
					left = left[i:]
					left = strings.ReplaceAll(left," ","")
					_,ok := verts[left]
					if ok {
						fmt.Println(left+":","cant redeclare with typename annotation")
					}
					right = strings.ReplaceAll(right," ","")
					if strings.ContainsAny(left,special+","+"[]") {
						fmt.Println(left+":","cant contain any special characters")
						continue
					}
					if right == "" {
						
 						right = "[0.5,0.5,0],[0,0.5,0],[-0.5,0.5,0]"
					}
					v,err := verts.valueof(right)
					if err != nil {
						fmt.Println(right+":",err)
						continue
					}
					var v2 variable
					switch(typename){
						case "triangle":
							v2,err = v.astriangle()
						case "point":
							v2,err = v.aspoint()
						case "line":
	 						v2,err = v.asline()
	
					}
					if err != nil {
						fmt.Println(stuff+":",err)
						continue
					}
					verts[left] = v2
					names = append(names,left)
				}
				
			default:
				
		}
		window.SwapBuffers()
		glfw.PollEvents()
		
	}
}
