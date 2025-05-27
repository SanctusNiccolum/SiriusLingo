import React from 'react';
import './Register.css';

function Register() {
    return (
        <div className="container">
            <h2>Register</h2>
            <form>
                <div className="form-group">
                    <label htmlFor="username">Username:</label>
                    <input type="text" id="username" name="username" required />
                </div>
                <div className="form-group">
                    <label htmlFor="email">Email:</label>
                    <input type="email" id="email" name="email" required />
                </div>
                <div className="form-group">
                    <label htmlFor="password">Password:</label>
                    <input type="password" id="password" name="password" required />
                </div>
                <button type="submit">Register</button>
            </form>
            <div className="link">Already have an account? <a href="/login">Login</a></div>
        </div>
    );
}

export default Register;