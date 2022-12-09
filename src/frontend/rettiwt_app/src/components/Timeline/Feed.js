import React from "react";
import "../../App.css";
import { Stack } from "react-bootstrap";
import FeedPub from "./FeedPub";

function Feed() {
    return (
        <Stack className="w-75 d-flex justify-content-center mx-auto" gap={1} direction='column-reverse'>
            <FeedPub />
            <FeedPub />
            <FeedPub />
            <FeedPub />
            <FeedPub />
        </Stack>
    );
}


export default Feed;