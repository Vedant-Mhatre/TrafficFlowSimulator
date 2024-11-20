#include <iostream>
#include <vector>
#include <chrono>
#include <thread>

const int GRID_WIDTH = 20;
const int GRID_HEIGHT = 10;
const int SIMULATION_STEPS = 50;

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

void clearScreen()
{
#ifdef _WIN32
    system("cls");
#else
    system("clear");
#endif
}

void moveVehicles(std::vector<Vehicle> &vehicles)
{
    for (auto &v : vehicles)
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

int main()
{
    // Initialize vehicles
    std::vector<Vehicle> vehicles = {
        {GRID_WIDTH / 2, GRID_HEIGHT - 1, UP},
        {0, GRID_HEIGHT / 2, RIGHT}};

    for (int step = 0; step < SIMULATION_STEPS; ++step)
    {
        clearScreen();
        std::cout << "Step: " << step + 1 << std::endl;

        // Move vehicles
        moveVehicles(vehicles);

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
