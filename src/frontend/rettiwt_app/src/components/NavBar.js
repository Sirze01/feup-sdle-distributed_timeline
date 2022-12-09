import { Navbar, Container, Form, Nav } from "react-bootstrap"
import logo from '../logo.svg';
import '../App.css';

function NavBar() {
  //change navbar according to token
  /*if (!token) {
  return (
    <>
      <Navbar bg="dark" variant="dark">
        <Container>
          <Navbar.Brand>

            <h2 className="my-primary"><a href="/"><img
              alt="logo"
              src={logo}
              color="#E25E0D"
              width="30"
              height="30"
              className="d-inline-block align-center App-logo-flipped mx-2"
            />
            </a>
              Rettiwt</h2>

          </Navbar.Brand>
        </Container>
      </Navbar>
    </>
  );
  }*/ //else {
  return (
    <>
      <Navbar bg="dark" variant="dark">
        <Container>
          <Navbar.Brand>

            <h2 className="my-primary"><a href="/"><img
              alt="logo"
              src={logo}
              color="#E25E0D"
              width="30"
              height="30"
              className="d-inline-block align-center App-logo-flipped mx-2"
            />
            </a>
              Rettiwt</h2>

          </Navbar.Brand>
         
          <Nav>
          <Form className="mx-3">
          <Form.Control
            type="search"
            placeholder="Search"
            className="me-2"
            aria-label="Search"
          />
        </Form>
            <Nav.Link className="text-light" href="/feed">Feed</Nav.Link>
            <Nav.Link className="text-light" href="/profile">My Profile</Nav.Link>
            <Nav.Link className="text-light mx-5" href="/">Log Out</Nav.Link>
          </Nav>
        </Container>
      </Navbar>
    </>
  );
  //}
}

export default NavBar;