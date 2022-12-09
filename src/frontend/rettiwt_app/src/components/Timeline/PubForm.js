import React from 'react';
import '../../App.css';
import { Form } from 'react-bootstrap';

function PubForm() {

    return (
        <Form className='Publish-form gap-2'>

            <Form.Group className='w-50' controlId="ControlTextarea1">
                <Form.Label>What is on your mind?</Form.Label>
                <Form.Control as="textarea" rows={3} />
            </Form.Group>
            
        </Form>
    );

}

export default PubForm;