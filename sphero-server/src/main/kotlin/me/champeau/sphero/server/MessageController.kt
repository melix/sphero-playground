package me.champeau.sphero.server

import io.micronaut.http.annotation.Controller
import io.micronaut.http.annotation.Get
import io.micronaut.http.HttpStatus
import io.micronaut.http.annotation.Post
import javax.inject.Inject

@Controller("/message")
class MessageController(@Inject val producer: MessageProducer) {
    @Post(uri="/")
    fun sendMessage(color: String?, speed: String?, angle: String?) {
        if (color != null) {
            producer.sendMessage("color", color)
        }
        if (speed != null) {
            producer.sendMessage("speed", speed)
        }
        if (angle != null) {
            producer.sendMessage("angle", angle)
        }
    }
}