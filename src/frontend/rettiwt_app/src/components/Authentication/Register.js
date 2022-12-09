import React, { useState } from 'react';
import '../../App.css';
import NavBar from '../NavBar';
import { Button, Form } from 'react-bootstrap';
import PropTypes from 'prop-types';

function Register(setToken) {

    /*    async function registerUser(credentials) {
           let url = `http://localhost:3000/register`;
           console.log(JSON.stringify(credentials))
           return fetch(url, {
               method: 'POST',
               headers: {
                   'Content-Type': 'application/json'
               },
               body: JSON.stringify(credentials)
           })
               .then(data => data.json())
               .catch(err => {
                   console.log(err)
               })
       } */
    async function registerUser(credentials) {
        console.log("Registing in user...");
        console.log(credentials);
    }
    const [user, setUser] = useState({
        username: "",
        password: ""
    })
    const [errors, setErrors] = useState({})



    const setUserInput = (event) => {
        const field = event.target.name
        const value = event.target.value

        setUser({ ...user, [field]: value })

        if (!!errors[field]) setErrors({
            ...errors,
            [field]: null
        })
    }

    const handleSubmit = async e => {

        e.preventDefault();
        //Clean all previous errors
        setErrors({})

        const token = await registerUser(user);

        if (!token) {
            setErrors({ server: "Server is down" })
            return
        }

        switch (token.code) {
            case 200:
                setToken(token.token);
                window.location.replace("/")
                break;
            case 400:
                setErrors({ server: token.message })
                break;
            case 401:
            case 402:
                setErrors({ username: token.message })
                break;
            case 403:
                setErrors({ password: token.message })
                break;

            default:
                console.log("Something went wrong")
        }

        // set token after receiving
    }
    return (
        <>
            <NavBar></NavBar>
            <Form onSubmit={handleSubmit} className="App Login-form gap-2">
                <h1 className='Login-header my-primary'>Register</h1>
                <Form.Group className="mb-2" controlId="validationCustomUsername" >
                    <Form.Control type="text" size="lg" name="username" placeholder="Username" onChange={setUserInput}  required />
                    <Form.Control.Feedback type="invalid">
                        Please choose a username.
                    </Form.Control.Feedback>

                </Form.Group>
                <Form.Group className="mb-2" controlId="formBasicPassword" >
                    <Form.Control size="lg" type="password" name="password" placeholder="Password" onChange={setUserInput}  required />
                    <Form.Control.Feedback type="invalid">
                        Please enter a valid password
                    </Form.Control.Feedback>

                </Form.Group>
                <Button variant="secondary" size="lg" type="submit" >Sign up</Button>
                <Form.Text className="text-muted">
                    Already have an account? <a href="/login">Sign in</a>
                </Form.Text>
            </Form>
        </>
    );
}
Register.propTypes = {
    setToken: PropTypes.func.isRequired
}
export default Register;
