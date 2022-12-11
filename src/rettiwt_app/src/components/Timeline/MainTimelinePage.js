import React from 'react';
import '../../App.css';
import NavBar from '../NavBar';
import PubForm from './PubForm'
import Feed from './Feed'




function MainTimelinePage() {

    return (
        <>
            < NavBar ></NavBar >
                <PubForm></PubForm>
                <Feed></Feed>
        </>
    );
}

export default MainTimelinePage;
