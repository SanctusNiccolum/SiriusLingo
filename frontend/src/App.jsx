import React from 'react';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';
import Index from "./components/Index.jsx";

function App() {
    return (
        <Router>
            <Routes>
                <Route path="/main" element={<Index />} />
                <Route path="/" element={<Index />} />
            </Routes>
        </Router>
    );
}

export default App;