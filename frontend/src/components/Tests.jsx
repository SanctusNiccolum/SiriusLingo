import React, { useState } from 'react';
import './Tests.css';

function LegalTest() {
    const [currentQuestion, setCurrentQuestion] = useState(1);
    const [showEtymology, setShowEtymology] = useState(false);
    const [selectedOption, setSelectedOption] = useState(null);

    const questions = [
        {
            id: 1,
            text: 'Select the correct translation for the word "peka":',
            options: ['Bakery', 'River', 'Mountain', 'Street', 'House'],
            etymology: 'The word "peka" comes from the Proto-Slavic *peka, meaning "baking" or "oven".'
        }
        // Можно добавить больше вопросов
    ];

    const handleSubmit = () => {
        if (selectedOption !== null) {
            setShowEtymology(true);
        }
    };

    const handleNext = () => {
        setSelectedOption(null);
        setShowEtymology(false);
        setCurrentQuestion(prev => prev + 1);
    };

    return (
        <div className="app-container">
            <header className="app-header">
                <h1>Lingua</h1>
                <button className="finish-button">Finish Test</button>
            </header>

            {/* Основное содержимое */}
            <div className="content-wrapper">
                <div className="question-container">
                    <h2>Question {currentQuestion}</h2>
                    <p className="question-text">{questions[0].text}</p>

                    <div className="options-list">
                        {questions[0].options.map((option, index) => (
                            <label key={index} className="option-item">
                                <span>{option}</span>
                                <input className="radio-item"
                                    type="radio"
                                    name="translation-option"
                                    checked={selectedOption === index}
                                    onChange={() => setSelectedOption(index)}
                                />
                            </label>
                        ))}
                    </div>

                    <div className="action-buttons">
                        {!showEtymology ? (
                            <button
                                className="submit-btn"
                                onClick={handleSubmit}
                                disabled={selectedOption === null}
                            >
                                Submit
                            </button>
                        ) : (
                            <button className="next-btn" onClick={handleNext}>
                                Next ➜
                            </button>
                        )}
                    </div>
                </div>

                {/* Etymology - фиксированная внизу */}
                {showEtymology && (
                    <div className="etymology-container">
                        <div className="etymology-content">
                            <h3>Etymology</h3>
                            <p>{questions[0].etymology}</p>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
}

export default LegalTest;