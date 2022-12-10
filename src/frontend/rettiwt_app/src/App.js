import './App.css';
import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import WelcomePage from './components/WelcomePage';
import Login from './components/Authentication/Login';
import Register from './components/Authentication/Register';
import MainTimelinePage from './components/Timeline/MainTimelinePage';
import ProfilePage from './components/Timeline/Profile/ProfilePage';
import useToken from './components/Authentication/useToken';


function App() {

  const { token, setToken } = useToken();

  if (!token)
    return (
      <Router>
        <Routes>
          <Route exact path='/' element={<WelcomePage />} />
          <Route path='/login' element={<Login setToken={setToken} />} />
          <Route path='/register' element={<Register setToken={setToken} />} />
          <Route path="/feed" element={<MainTimelinePage />} />
          <Route path="/profile" element={<ProfilePage />} />

        </Routes>
      </Router>
    )
  else
    return(<Router>
      <Routes>
        <Route exact path='/' element={<WelcomePage />} />
        <Route path='/login' element={<Login setToken={setToken} />} />
        <Route path='/register' element={<Register setToken={setToken} />} />
        <Route path="/feed" element={<MainTimelinePage setToken={setToken}/>} />
        <Route path="/profile" element={<ProfilePage setToken={setToken}/>} />

        </Routes>
    </Router>)
}





export default App;
//        <Route path='/login' element={<Login/>}/>
