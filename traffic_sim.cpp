#include <iostream>
#include <vector>
#include <chrono>
#include <thread>

const int GRID_WIDTH = 20;
const int GRID_HEIGHT = 10;
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

void moveVehicles(std::vector<Vehicle> &vehicles, TrafficLight &light)
{
    for (auto &v : vehicles)
    {
        bool canMove = true;

        if (v.x == GRID_WIDTH / 2 && v.y == GRID_HEIGHT / 2)
        {
            if ((v.dir == UP || v.dir == DOWN) && !light.isGreen)
                canMove = false;
            if ((v.dir == LEFT || v.dir == RIGHT) && light.isGreen)
                canMove = false;
        }

        if (canMove)
        {
            switch (v.dir)
            {
            case UP:
                if (v.y > 0)
                    v.y--;
                break;
            case DOWN:
                if (v.y < GRID_HEIGHT - 1)
                    v.y++;
                break;
            case LEFT:
                if (v.x > 0)
                    v.x--;
                break;
            case RIGHT:
                if (v.x < GRID_WIDTH - 1)
                    v.x++;
                break;
            }
        }
    }
}

void displayGrid(std::vector<Vehicle> &vehicles, TrafficLight &light)
{
    char grid[GRID_HEIGHT][GRID_WIDTH];
    for (int i = 0; i < GRID_HEIGHT; ++i)
        for (int j = 0; j < GRID_WIDTH; ++j)
            grid[i][j] = ' ';

    for (int i = 0; i < GRID_HEIGHT; ++i)
        grid[i][GRID_WIDTH / 2] = '|';
    for (int j = 0; j < GRID_WIDTH; ++j)
        grid[GRID_HEIGHT / 2][j] = '-';

    grid[GRID_HEIGHT / 2][GRID_WIDTH / 2] = light.isGreen ? 'G' : 'R';

    for (const auto &v : vehicles)
        grid[v.y][v.x] = 'V';

    for (int i = 0; i < GRID_HEIGHT; ++i)
    {
        for (int j = 0; j < GRID_WIDTH; ++j)
            std::cout << grid[i][j];
        std::cout << std::endl;
    }
}

int main()
{
    std::vector<Vehicle> vehicles = {
        {GRID_WIDTH / 2, GRID_HEIGHT - 1, UP},
        {0, GRID_HEIGHT / 2, RIGHT}};

    TrafficLight light = {true, 0}; // Vertical green to start

    for (int step = 0; step < SIMULATION_STEPS; ++step)
    {
        clearScreen();
        std::cout << "Step: " << step + 1 << std::endl;

        // Update traffic light
        updateTrafficLight(light);
        std::cout << "Traffic Light: " << (light.isGreen ? "Green (Vertical)" : "Red (Horizontal)") << std::endl;

        // Display vehicle positions
        for (const auto &v : vehicles)
        {
            std::cout << "Vehicle at (" << v.x << ", " << v.y << ")\n";
        }

        std::this_thread::sleep_for(std::chrono::milliseconds(500));
    }

    std::cout << "Simulation Complete!" << std::endl;
    return 0;
}
