import React from 'react';
import logo from '../logo.svg';
import '../App.css';
import NavBar from './NavBar';
import { Button } from 'react-bootstrap';




function WelcomePage() {
  return (
    <>
    <NavBar>    </NavBar>
    <div className="App">
      <header className="App-header">
        <img src={logo} className="App-logo App-logo-flipped" alt="logo" />
        <h1 className="my-primary">
          Rettiwt
        </h1>
        <Button href="/register" variant="secondary" className="mt-5 mb-2" size="lg">
          Register
        </Button>
        <p className='my-text'>If you already have an account <a href="/login">Login</a></p>

      </header>
    </div>
    </>
  );
}

export default WelcomePage;
