package me.champeau.sphero.server

import io.micronaut.configuration.kafka.annotation.KafkaClient
import io.micronaut.configuration.kafka.annotation.KafkaKey
import io.micronaut.configuration.kafka.annotation.Topic


@KafkaClient
interface MessageProducer {
    @Topic("test")
    fun sendMessage(@KafkaKey key: String, message: String)


}