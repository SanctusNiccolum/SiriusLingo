import React from 'react';
import './Index.css';

function Index() {
    return (
        <div className="container">
            <div className="welcome"><span className="emoji">🧑</span> Welcome, [username]</div>
            <div className="stats">
                <div><span className="emoji">⭐️</span> Your total score: [score]</div>
                <div><span className="emoji">❓</span> Completed questions: [count]</div>
            </div>
            <div className="site-info">
                <h3>LinguaTest</h3>
                <p>LinguaTest is your ultimate tool for mastering language skills through interactive tests and deep etymology insights. Whether you're a beginner or an advanced learner, our platform helps you grow by making language learning engaging and insightful.</p>
                <p><strong>How LinguaTest can help you:</strong></p>
                <ul>
                    <li><strong>Expand your vocabulary:</strong> Learn new words and their meanings through carefully designed tests.</li>
                    <li><strong>Understand word origins:</strong> Dive into the etymology of words to better grasp their history and usage.</li>
                    <li><strong>Improve language comprehension:</strong> Practice with real-world examples to enhance your understanding of context and usage.</li>
                    <li><strong>Boost memory retention:</strong> Interactive tests help reinforce your learning, making it easier to remember new words.</li>
                    <li><strong>Track your progress:</strong> Monitor your scores and completed questions to see your improvement over time.</li>
                </ul>
            </div>
            <a href="/tests" className="button">Start Test</a>
        </div>
    );
}

export default Index;