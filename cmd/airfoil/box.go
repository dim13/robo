package main

import "github.com/dim13/robo"

var Box = robo.Path{
	robo.Point{0.0, 0.0},
	robo.Point{-0.1, 0.0},
	robo.Point{-0.1, -0.1},
	robo.Point{1.1, -0.1},
	robo.Point{1.1, 0.0},
	robo.Point{1.0, 0.0},
}

var DrillHoles = []robo.Path{
	robo.Path{robo.Point{0.0, 0.0}, robo.Point{-0.1, -0.1}},
	robo.Path{robo.Point{-0.1, 0.0}, robo.Point{0.0, -0.1}},
	robo.Path{robo.Point{1.0, 0.0}, robo.Point{1.1, -0.1}},
	robo.Path{robo.Point{1.1, 0.0}, robo.Point{1.0, -0.1}},
}

var Marks = []robo.Path{
	robo.Path{robo.Point{0.0, -0.2}, robo.Point{0.0, 0.2}},
	robo.Path{robo.Point{0.1, -0.2}, robo.Point{0.1, 0.2}},
	robo.Path{robo.Point{0.2, -0.2}, robo.Point{0.2, 0.2}},
	robo.Path{robo.Point{0.3, -0.2}, robo.Point{0.3, 0.2}},
	robo.Path{robo.Point{0.4, -0.2}, robo.Point{0.4, 0.2}},
	robo.Path{robo.Point{0.5, -0.2}, robo.Point{0.5, 0.2}},
	robo.Path{robo.Point{0.6, -0.2}, robo.Point{0.6, 0.2}},
	robo.Path{robo.Point{0.7, -0.2}, robo.Point{0.7, 0.2}},
	robo.Path{robo.Point{0.8, -0.2}, robo.Point{0.8, 0.2}},
	robo.Path{robo.Point{0.9, -0.2}, robo.Point{0.9, 0.2}},
	robo.Path{robo.Point{1.0, -0.2}, robo.Point{1.0, 0.2}},
}
