import React from "react";
import "../../App.css";
import { Card, Stack } from "react-bootstrap";

function FeedPub() {
    return (

        <Card border="dark" className="m-4 overflow-hidden">
            <Card.Body>
                <Stack gap={1} direction="horizontal">
                    <h5 className="mx-3"><a href="/feed"> Username</a></h5>
                    <h5 className="font-weight-bold" >Time Ago</h5>
                </Stack>
                <p className="text-wrap mx-3">Hello This is a test Message</p>
            </Card.Body>
        </Card>

    );
}


export default FeedPub;
;
