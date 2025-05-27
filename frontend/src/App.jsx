import React from 'react';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import Register from './components/Register';
import Login from './components/Login';
import Index from './components/Index';
import LegalTest from "./components/Tests.jsx";

function App() {
    return (
        <Router>
            <Routes>
                <Route path="/register" element={<Register />} />
                <Route path="/login" element={<Login />} />
                <Route path="/main" element={<Index />} />
                <Route path="/tests" element={<LegalTest />} />
                <Route path="/" element={<Index />} />
            </Routes>
        </Router>
    );
}

export default App;