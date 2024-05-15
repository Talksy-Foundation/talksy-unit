const talksySFU = {
    socket: WebSocket,
    pc: RTCPeerConnection,
    room_id: String,
    self_id: String
}


talksySFU.connect = () => {
    console.log("Connected");

    return new WebSocket('ws://localhost:8080')
}
talksySFU.makeId = (length) => {
    let result = '';
    const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    const charactersLength = characters.length;
    let counter = 0;
    while (counter < length) {
        result += characters.charAt(Math.floor(Math.random() * charactersLength));
        counter += 1;
    }
    return result;
}
talksySFU.sendOffer = async (socket, pc, self_id, room_id) => {
    const offer = await pc.createOffer();
    await pc.setLocalDescription(offer);
    const message= {
        self_id: self_id,
        room_id: room_id,
        offer: offer
    }
    console.log('Offer created: ', message);

    socket.send(JSON.stringify({event: 'offer', data: message}));
}


talksySFU.joinRoom = async () => {
    const message= {
        room_id: talksySFU.room_id,
        self_id: talksySFU.self_id
    };
    talksySFU.socket.send(JSON.stringify({ event: 'joinRoom', data: message }));
}

talksySFU.leaveRoom = async () => {
    const message= {
        room_id: talksySFU.room_id,
        self_id: talksySFU.self_id
    }
    talksySFU.socket.send(JSON.stringify({ event: 'leaveRoom', data: message }));
}


talksySFU.setup = () => {
    talksySFU.socket.onmessage = async (event) => {
        const data = JSON.parse(event.data.toString());
        switch (data.event) {
            case 'answer': {
                const value = JSON.parse(data.data);
                talksySFU.pc.setRemoteDescription(value);
                talksySFU.pc.oniceconnectionstatechange = (event) => {
                    console.log(event);
                    console.log(talksySFU.pc.iceConnectionState);
                };
                talksySFU.pc.onicegatheringstatechange = (event) => {
                    console.log(event);
                    console.log(talksySFU.pc.iceConnectionState);
                }
                break;
            }
            case 'candidate': {
                const value = JSON.parse(data.data);
                await talksySFU.pc.addIceCandidate(new RTCIceCandidate(value));
                break;
            }
            case 'offer': {
                const value = JSON.parse(data.data);
                console.log("Offer ", value);
                await talksySFU.pc.setRemoteDescription(value);
                await talksySFU.pc.createAnswer().then(
                    async (answer) => {
                        await talksySFU.pc.setLocalDescription(answer);
                        const message = {
                            self_id: value.self_id,
                            room_id: value.room_id,
                            answer: answer
                        }
                        talksySFU.socket.send(JSON.stringify({ event: 'answer', data: message }));
                    }
                );
                break;
            }
        }
    };

    talksySFU.pc.onicecandidate = (event) => {
        event.candidate ? talksySFU.socket.send(JSON.stringify({ event: 'ice-candidate', data: { room_id: talksySFU.room_id, self_id: talksySFU.self_id, candidate: event.candidate } })) : null;
    };
};




talksySFU.init = async (room_id) => {
    talksySFU.room_id = room_id
    talksySFU.self_id = talksySFU.makeId(10)
    talksySFU.socket = talksySFU.connect()
    await talksySFU.joinRoom()
    talksySFU.pc = new RTCPeerConnection()
    talksySFU.setup()
    await talksySFU.sendOffer()
}