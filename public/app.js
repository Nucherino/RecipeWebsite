document.addEventListener('DOMContentLoaded', function() {
    fetchRecipes();
});

function fetchRecipes() {
    fetch('/recipes')
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.json();
        })
        .then(recipes => {
            const recipeList = document.getElementById('recipe-list');
            recipeList.innerHTML = ''; 
            recipes.forEach(recipe => {
                const recipeItem = document.createElement('div');
                recipeItem.className = 'recipe-item'; 
                recipeItem.innerHTML = `
                    <h3>${recipe.name}</h3>
                    <p><strong>Ingredients:</strong> ${recipe.ingredients.join(', ')}</p>
                    <p><strong>Instructions:</strong> ${recipe.instructions}</p>
                `;
                recipeList.appendChild(recipeItem);
            });
        })
        .catch(error => {
            console.error('Error fetching recipes:', error);
        });
}

document.getElementById('recipe-form').addEventListener('submit', function(event) {
    event.preventDefault(); 
    const name = document.getElementById('name').value;
    const ingredients = document.getElementById('ingredients').value.split(',').map(ing => ing.trim()); 
    const instructions = document.getElementById('instructions').value;

    const recipe = {
        name: name,
        ingredients: ingredients,
        instructions: instructions
    };

    fetch('/recipes', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(recipe)
    })
    .then(response => response.json())
    .then(data => {
        console.log('Success:', data);
        fetchRecipes();
    })
    .catch((error) => {
        console.error('Error:', error);
    });
});
