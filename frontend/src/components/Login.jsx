import React from 'react';
import './Login.css';

function Login() {
    return (
        <div className="container">
            <h2>Login</h2>
            <form>
                <div className="form-group">
                    <label htmlFor="username">Username:</label>
                    <input type="text" id="username" name="username" required />
                </div>
                <div className="form-group">
                    <label htmlFor="password">Password:</label>
                    <input type="password" id="password" name="password" required />
                </div>
                <button type="submit">Login</button>
            </form>
            <div className="link">Don't have an account? <a href="/register">Register</a></div>
        </div>
    );
}

export default Login;