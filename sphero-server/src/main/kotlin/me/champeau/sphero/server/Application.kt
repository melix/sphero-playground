package me.champeau.sphero.server

import io.micronaut.runtime.Micronaut.*
fun main(args: Array<String>) {
	build()
	    .args(*args)
		.packages("me.champeau.sphero.server")
		.start()
}

