#include <iostream>
#include <vector>
#include <chrono>
#include <thread>

const int GRID_WIDTH = 20;
const int GRID_HEIGHT = 10;
const int NUM_VEHICLES = 2;
const int SIMULATION_STEPS = 50;
const int TRAFFIC_LIGHT_INTERVAL = 5;

enum Direction
{
    UP,
    DOWN,
    LEFT,
    RIGHT
};

struct Vehicle
{
    int x, y;
    Direction dir;
};

struct TrafficLight
{
    bool isGreen;
    int timer;
};

void clearScreen()
{
#ifdef _WIN32
    system("cls");
#else
    system("clear");
#endif
}

void updateTrafficLight(TrafficLight &light)
{
    light.timer++;
    if (light.timer >= TRAFFIC_LIGHT_INTERVAL)
    {
        light.isGreen = !light.isGreen;
        light.timer = 0;
    }
}

void moveVehicles(std::vector<Vehicle> &vehicles, TrafficLight &light, int &totalSpeed)
{
    for (auto &v : vehicles)
    {
        bool canMove = true;

        // Calculate next position based on direction
        int nextX = v.x;
        int nextY = v.y;

        switch (v.dir)
        {
        case UP:
            nextY = v.y - 1;
            break;
        case DOWN:
            nextY = v.y + 1;
            break;
        case LEFT:
            nextX = v.x - 1;
            break;
        case RIGHT:
            nextX = v.x + 1;
            break;
        }

        // Check boundaries
        if (nextX < 0 || nextX >= GRID_WIDTH || nextY < 0 || nextY >= GRID_HEIGHT)
        {
            canMove = false; // Prevent moving out of the grid
        }

        // Check for traffic light before entering the intersection
        if (canMove && nextX == GRID_WIDTH / 2 && nextY == GRID_HEIGHT / 2)
        {
            if ((v.dir == UP || v.dir == DOWN) && !light.isGreen)
            {
                canMove = false; // Red light for vertical movement
            }
            if ((v.dir == LEFT || v.dir == RIGHT) && light.isGreen)
            {
                canMove = false; // Red light for horizontal movement
            }
        }

        // Move vehicle if possible
        if (canMove)
        {
            v.x = nextX;
            v.y = nextY;
            totalSpeed++;
        }
    }
}

void displayGrid(std::vector<Vehicle> &vehicles, TrafficLight &light)
{
    char grid[GRID_HEIGHT][GRID_WIDTH];

    // Initialize grid
    for (int i = 0; i < GRID_HEIGHT; ++i)
        for (int j = 0; j < GRID_WIDTH; ++j)
            grid[i][j] = ' ';

    // Place roads (simple cross intersection)
    for (int i = 0; i < GRID_HEIGHT; ++i)
        grid[i][GRID_WIDTH / 2] = '|';

    for (int j = 0; j < GRID_WIDTH; ++j)
        grid[GRID_HEIGHT / 2][j] = '-';

    grid[GRID_HEIGHT / 2][GRID_WIDTH / 2] = light.isGreen ? 'G' : 'R';

    // Place vehicles
    for (auto &v : vehicles)
    {
        grid[v.y][v.x] = 'V';
    }

    // Display grid
    for (int i = 0; i < GRID_HEIGHT; ++i)
    {
        for (int j = 0; j < GRID_WIDTH; ++j)
        {
            std::cout << grid[i][j];
        }
        std::cout << std::endl;
    }
}

int main()
{
    // Initialize vehicles
    std::vector<Vehicle> vehicles;
    for (int i = 0; i < NUM_VEHICLES; ++i)
    {
        Vehicle v;
        if (i % 2 == 0)
        {
            // Vertical vehicles
            v.x = GRID_WIDTH / 2;
            v.y = GRID_HEIGHT - 1;
            v.dir = UP;
        }
        else
        {
            // Horizontal vehicles
            v.x = 0;
            v.y = GRID_HEIGHT / 2;
            v.dir = RIGHT;
        }
        vehicles.push_back(v);
    }

    // Initialize traffic light at intersection
    TrafficLight light = {true, 0}; // Vertical green to start

    int totalSpeed = 0;

    // Simulation loop
    for (int step = 0; step < SIMULATION_STEPS; ++step)
    {
        clearScreen();
        std::cout << "Step: " << step + 1 << std::endl;

        // Display grid
        displayGrid(vehicles, light);

        // Move vehicles and update traffic light
        moveVehicles(vehicles, light, totalSpeed);
        updateTrafficLight(light);

        // Pause for a moment
        std::this_thread::sleep_for(std::chrono::milliseconds(500));
    }

    // Calculate performance metrics
    double averageSpeed = static_cast<double>(totalSpeed) / (NUM_VEHICLES * SIMULATION_STEPS);
    std::cout << "\nSimulation Complete!" << std::endl;
    std::cout << "Average Vehicle Speed: " << averageSpeed << " units/step" << std::endl;

    return 0;
}
