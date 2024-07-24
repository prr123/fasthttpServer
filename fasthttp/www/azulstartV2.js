window.onload = (event) => {
    const azul = new HtmlPage;
    let pageObj = {
        mainDiv: {
            style: {
                width: 'calc(100% - 300px)',
                border: '1px solid blue',
                minHeight: '300px',
            },
            id: 'divMain',
        },
        header: {
            style: {
                height: '100px',
                margin: '10px',
                border: '1px solid DeepPink',
                position: 'relative',
            },
            id: 'header',
            className: 'pagSections',
        },
        section: {
            style: {
                minHeight: '500px',
                margin: '10px',
                border: '1px solid Tomato',
                position: 'relative',
            },
            id: 'docmain',
            className: 'pagSections',
        },
        footer: {
            style: {
                height: '100px',
                margin: '10px',
                border: '1px solid green',
                position: 'relative',
            },
            id: 'footer',
            className: 'pagSections',
        },
    }

    let metaObj = {
        metaNames: [
            {name: 'description', content: 'a blog'},
            {name: 'author', content: 'prr'},
            {name: 'date', content: '1/10/2022'},
            ],
    };

/*
    let linkObj = {
        id: 'azulCss',
        type: 'text/css',
        href: 'azulLib.css',
    }
*/
    azul.init(pageObj);

    azul.addMeta(metaObj);

//    azul.addLink(linkObj);

    azul.addStyleObj();

//    azul.addScript('azulLib.js');


    let hdStyl = {
        color: 'MediumPurple',
        margin: 'auto',
        textAlign: 'center',
        padding: '0.5em',
    };

    let hdObj = {
        style: hdStyl,
        parent: azul.header,
        id: 'header',
        className: 'doch3',
        textContent: 'testing websocket',
        typ: 'h3',
    }
    azul.addElement(hdObj);

    let hd2Obj = {
        style: hdStyl,
        parent: azul.docbody,
        id: 'docmain',
        className: 'doch3',
        textContent: 'docmain',
        typ: 'h3',
    }
    azul.addElement(hd2Obj);

    let hdfooterObj = {
        style: hdStyl,
        parent: azul.footer,
        id: 'footer',
        className: 'doch3',
        textContent: 'footer',
        typ: 'h3',
    }
    azul.addElement(hdfooterObj);

    let butObj = {
        style: {
			height: '30px',
			width: '100px',
			position: 'absolute',
			top: '50px',
			left: '50px',
			border: '1px solid green',
        },
        parent: azul.header,
        typ: 'button',
		textContent: 'press',
		cpar: 'tstpar',
		butFunc() {
			let count = 1;
//			let res = false
			var webSocket = new WebSocket('ws://89.116.30.49:9005/hijack');
			webSocket.binaryType = "arraybuffer";

			webSocket.addEventListener("error", (event) => {
  				console.log("WebSocket error: ", event);
			});
			webSocket.addEventListener("open", (event) => {
				console.log('socket open');
  				webSocket.send("Hello Server!");
				console.log('sent message');
			});
			webSocket.addEventListener("message", (event) => {
				if (event.data instanceof ArrayBuffer) {
    			// binary frame
					var barray = new Uint8Array(event.data);
					console.log("bin msg:" + barray.length);
					for (var i = 0; i < barray.length; i++) {
    					console.log(i+": " + barray[i]);
					}
					console.log("End of binary message");  
    				const view = new DataView(event.data);
    				console.log("binary msg rec>", view.getInt32(0, true));
					const buffer = new ArrayBuffer(4);
					const outview = new DataView(buffer);
					let num = view.getInt32(0, true);
					rval = num + 3;
					console.log("num: " + num + " rval: " + rval);
					outview.setInt32(0, rval, true)
	  				webSocket.send(buffer);
  				} else {
					console.log("Message from server> ", event.data);
					if (event.data == 'end') {
						webSocket.close();
						return;
					}
	  				webSocket.send('Hello Server ' + count + '!');
					count++;
				}
			});
		},
    };
    azul.addButton(butObj);

//	let cfval = butObj['cfun'] !== undefined
//	let dfval = butObj['dfun'] !== undefined

//	let res = butObj.cfun('jim');
//	console.log('hello: ' + res);
    document.body.appendChild(azul.divMain);
};


/*
webSocket.onopen = function(e) {
	alert("[open] Connection established");
	alert("Sending to server");
    socket.send("My name is John");
};
 
*/
