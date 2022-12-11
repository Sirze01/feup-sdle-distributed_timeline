import React from 'react';
import '../../../App.css';
import NavBar from '../../NavBar';
import Feed from '../Feed'
import ProfileSettings from './ProfileSettings'


function ProfilePage() {

    return (
        <>
            < NavBar ></NavBar >
            <ProfileSettings></ProfileSettings>
            <hr className="solid"></hr>
            <Feed></Feed>
        </>
    );
}

export default ProfilePage;