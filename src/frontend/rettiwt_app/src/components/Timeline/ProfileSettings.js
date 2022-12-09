import React from 'react';
import '../../App.css';
import { Button, Stack, Row, Col } from 'react-bootstrap';

function ProfileSettings() {

    let following = false;
    const handleFollow = () => {
        console.log("Following: username")
        following = true;
    }

    //verify if the user is the owner of the profile
    return (
            <Row className='mt-5 mx-5'>
                <Col md={{ span: 7, offset: 1 }}>
                    <h2 className='my-primary p-3'> Username<Button style={{ background: "#E25E0D", border: "#E25E0D" }} className="mx-3" onClick={handleFollow} size="medium">{following ? "Unfollow" : "Follow"}
                    </Button></h2>
                </Col>
                <Col className='my-auto'>
                    <Stack gap={3} direction="horizontal">
                        <h5 className="font-weight-bold">x Followers</h5>
                        <h5 className="font-weight-bold" >y Following</h5>
                    </Stack>
                </Col>

            </Row>
    )
        ;
}
export default ProfileSettings;
