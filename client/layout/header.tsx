import * as React from "react";
import {Link} from 'react-router-dom'

function Header() {
  return (
    <nav className="navbar" role="navigation" aria-label="main navigation">
      <div id="navbarBasicExample" className="navbar-menu">
        <div className="navbar-start">
          <a className="navbar-item" href="/">Home</a>
          <a className="navbar-item">Documentation</a>
        </div>
        <div className="navbar-end">
          <div className="navbar-item">
            <div className="buttons">
              <a className="button is-dark">
                <strong>Sign up</strong>
              </a>
              <a className="button is-light">Log in</a>
            </div>
          </div>
        </div>
      </div>
    </nav>
  );
}

export default Header;
