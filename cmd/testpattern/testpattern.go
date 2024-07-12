package main

var TestPattern = []string{
	"FN0",                                 // orientation protrait
	"&100,100,100",                        // factor
	"^0,0",                                // offset
	"\\0,0",                               // lower left
	"M510,637",                            // move
	"BZ1,510,637,439,637,383,580,383,510", // bezier
	"BZ1,383,510,383,439,439,383,510,383", // bezier
	"BZ1,510,383,580,383,637,439,637,510", // bezier
	"BZ1,637,510,637,580,580,637,510,637", // bezier
	"M764,764",                            // move
	"D256,764",                            // draw
	"D256,256",                            // draw
	"D764,256",                            // draw
	"D764,764",                            // draw
	"M2,510",                              // move
	"D1018,510",                           // draw
	"M510,1018",                           // move
	"D510,2",                              // draw
	"M0,0",                                // move to origin
}
