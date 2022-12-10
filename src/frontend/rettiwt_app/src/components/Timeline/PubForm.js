import React, { useState } from 'react';
import '../../App.css';
import { Form, Row, Col,Button } from 'react-bootstrap';

function PubForm() {

    const [message, setMessage] = useState("");


    const handlePublish = () => {
        console.log("Publishing: " + message)
    }

    return (
        <Form className='Publish-form gap-2 mt-3'>

            <Form.Group className='w-50' controlId="ControlTextarea1">
                <Form.Label>What is on your mind?</Form.Label>
                <Form.Control value={message} maxLength={256} as="textarea" rows={3} onChange={e => setMessage(e.target.value)} />
                <Row className='my-2'>
                    <Col>
                        <p>{message.length} / 256</p>
                    </Col>
                    <Col className ="mx-5" xs={1}>
                    <Button style={{ background: "#E25E0D", border: "#E25E0D" }} className="mx-3" onClick={handlePublish} size="small">Publish
                    </Button>
                    </Col>
                </Row>
            </Form.Group>
        </Form>
    );

}

export default PubForm;

